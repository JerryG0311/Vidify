package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/JerryG0311/Vidify/internal/pubsub"
	"github.com/JerryG0311/Vidify/internal/routing"
	"github.com/JerryG0311/Vidify/internal/storage"
	_ "github.com/mattn/go-sqlite3"
	amqp "github.com/rabbitmq/amqp091-go"
)

var db *sql.DB

func main() {
	// SETTING UP DATABASE CONNECTION
	var err error
	db, err = sql.Open("sqlite3", "./data/vidify.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	connString := os.Getenv("RABBITMQ_URL")
	if connString == "" {
		connString = "amqp://guest:guest@localhost:5672/"
	}

	var conn *amqp.Connection

	// RETRY LOOP: Try to connect 5 times with a 2-second pause between each
	for i := 0; i < 5; i++ {
		fmt.Printf("Connecting to RabbitMQ (attempt %d)... ", i+1)
		conn, err = amqp.Dial(connString)
		if err == nil {
			fmt.Println("Connected!")
			break
		}
		fmt.Printf("Failed: %v. Retrying in 2s...\n", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to RabbitMQ after retries: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	defer ch.Close()

	err = ch.ExchangeDeclare(
		routing.ExchangeVideoTopic,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	fmt.Println("Vidify Worker started. Waiting for video jobs...")

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangeVideoTopic,
		routing.VideoQueue,
		routing.VideoUploadKey,
		pubsub.SimpleQueueDurable,
		handlerVideoJob,
	)

	if err != nil {
		log.Fatalf("Worker failed to subscribe: %v", err)
	}

	select {}
}

func handlerVideoJob(job routing.VideoJob) pubsub.AckType {
	fmt.Printf(" Worker received job %s. Starting transcode...\n", job.ID)

	// 1. Prepare Local Paths ( Temporary storage inside the container)
	inputLocal := fmt.Sprintf("/tmp/%s_input.mp4", job.ID)
	thumbLocal := fmt.Sprintf("/tmp/%s_thumb.jpg", job.ID)
	outputLocal := fmt.Sprintf("/tmp/%s_processed.mp4", job.ID)

	// Clean up local files when done
	defer os.Remove(inputLocal)
	defer os.Remove(thumbLocal)
	defer os.Remove(outputLocal)

	// 2. Download from S3 to local
	if err := storage.DownloadFromS3(job.SourcePath, inputLocal); err != nil {
		log.Printf("Download failed for job %s: %v", job.ID, err)
		time.Sleep(5 * time.Second)
		return pubsub.NackRequeue
	}

	if _, err := db.Exec("UPDATE videos SET status = ? WHERE id = ?", "PROCESSING", job.ID); err != nil {
		log.Printf("Failed to update status to PROCESSING for job %s: %v", job.ID, err)
	}

	// 3. Generate Thumbnail
	thumbCmd := exec.Command("ffmpeg", "-y", "-i", inputLocal, "-ss", "00:00:01.000", "-vframes", "1", thumbLocal)
	thumbOutput, err := thumbCmd.CombinedOutput()
	if err != nil {
		log.Printf("Thumbnail generation failed for job %s: %v | ffmpeg output: %s", job.ID, err, string(thumbOutput))
	}

	// 4. Main Transcode
	transcodeCmd := exec.Command("ffmpeg", "-y", "-i", inputLocal, outputLocal)
	transcodeOutput, err := transcodeCmd.CombinedOutput()
	if err != nil {
		log.Printf("Transcode failed for job %s: %v | ffmpeg output: %s", job.ID, err, string(transcodeOutput))
		if _, dbErr := db.Exec("UPDATE videos SET status = ? WHERE id = ?", "FAILED", job.ID); dbErr != nil {
			log.Printf("Failed to update status to FAILED for job %s: %v", job.ID, dbErr)
		}
		return pubsub.NackDiscard
	}

	// 5. Upload Results Back to S3
	fmt.Printf("Transcoding complete. Uploading results to S3...\n")

	processedKey := fmt.Sprintf("%s_processed.mp4", job.ID)
	processedS3URL, err := storage.UploadFileToS3(processedKey, outputLocal)
	if err != nil {
		log.Printf("Processed video upload failed for job %s: %v", job.ID, err)
		if _, dbErr := db.Exec("UPDATE videos SET status = ? WHERE id = ?", "FAILED", job.ID); dbErr != nil {
			log.Printf("Failed to update status to FAILED after processed upload error for job %s: %v", job.ID, dbErr)
		}
		return pubsub.NackRequeue
	}

	autoThumbURL := ""
	thumbBytes, thumbReadErr := os.ReadFile(thumbLocal)
	if thumbReadErr != nil {
		log.Printf("Thumbnail file could not be read for job %s: %v", job.ID, thumbReadErr)
	} else if len(thumbBytes) == 0 {
		log.Printf("Thumbnail file was empty for job %s", job.ID)
	} else {
		thumbKey := fmt.Sprintf("%s_thumb.jpg", job.ID)
		thumbReader := bytes.NewReader(thumbBytes)
		thumbURL, thumbUploadErr := storage.UploadToS3(thumbKey, thumbReader)
		if thumbUploadErr != nil {
			log.Printf("Thumbnail upload failed for job %s: %v", job.ID, thumbUploadErr)
		} else {
			autoThumbURL = thumbURL
		}
	}

	query := `
			UPDATE videos
			SET status = 'COMPLETED',
				source_path = ?,
				thumbnail_url = COALESCE(NULLIF(thumbnail_url, ''), ?)
			WHERE id = ?
			`
	if _, err = db.Exec(query, processedS3URL, autoThumbURL, job.ID); err != nil {
		log.Printf("Final DB update error for job %s: %v", job.ID, err)
		return pubsub.NackRequeue
	}

	return pubsub.Ack
}
