package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/JerryG0311/Vidify/internal/pubsub"
	"github.com/JerryG0311/Vidify/internal/routing"
	"github.com/JerryG0311/Vidify/internal/storage"
	amqp "github.com/rabbitmq/amqp091-go"
)

type VideoData struct {
	ID           string
	Title        string
	Description  string
	Playlist     string
	SourcePath   string
	ThumbnailURL string
	Views        int
	CreatedAt    time.Time
	Status       string
}

type GalleryPageData struct {
	Videos []VideoData
}

func main() {

	var err error
	// -- Database  --

	db, err := sql.Open("sqlite3", "vidify.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	query := `CREATE TABLE IF NOT EXISTS videos (
		id TEXT PRIMARY KEY, 
		user_id TEXT,
		status TEXT,
		source_path TEXT,
		thumbnail_url TEXT,
		title TEXT,
		description TEXT,
		playlist	TEXT,
		views INTEGER DEFAULT 0,
		created_at DATETIME
	
	)`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// 1. Establishing connection

	connString := os.Getenv("RABBITMQ_URL")
	if connString == "" {
		connString = "amqp://guest@localhost:5672/"
	}

	var conn *amqp.Connection

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

	fmt.Println("Vidify API started. Connecting to RabbitMQ...")

	// 2. Creating a channel to declare Exchange
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// 3. Declare video topic exchange
	err = ch.ExchangeDeclare(
		routing.ExchangeVideoTopic, // name
		"topic",                    // type
		true,                       // durable
		false,                      // auto-delete
		false,                      // internal
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	// 3.5a Declare Dead Letter Exchange
	err = ch.ExchangeDeclare(
		routing.ExchangeVideoDLX,
		"fanout",
		true,  // durable
		false, // auto-deleted
		false,
		false,
		nil,
	)

	// 3.5b Declare the "Failed Jobs" queue
	_, err = ch.QueueDeclare(
		routing.VideoDLQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false,
		nil,
	)

	// 3.5c Bind the Failed Queue to the DLX
	err = ch.QueueBind(routing.VideoDLQueue, "", routing.ExchangeVideoDLX, false, nil)

	// ---- New Web Server Code ---

	// 1. Parse the file from the request ("video" is the key used in the curl command)

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		// --- 1. SHOW THE FORM (GET REQUEST) ---
		if r.Method == http.MethodGet {
			tmpl, err := template.ParseFiles("web/templates/upload.html")
			if err != nil {
				http.Error(w, "Upload template not found", http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, nil)
			return
		}

		// --- 2. PROCESS THE UPLOAD (POST REQUEST) ---
		if r.Method == http.MethodPost {
			// Optional: Increase max memory for large video uploads (32MB default -> 500MB)
			r.ParseMultipartForm(500 << 20)

			title := r.FormValue("title")
			description := r.FormValue("description")
			playlist := r.FormValue("playlist")

			file, header, err := r.FormFile("video")
			if err != nil {
				http.Error(w, "Failed to get video file", http.StatusBadRequest)
				return
			}
			defer file.Close()

			if title == "" {
				title = header.Filename
			}

			job := routing.VideoJob{
				ID:           fmt.Sprintf("vid-%d", time.Now().Unix()),
				SourcePath:   "",
				TargetFormat: "mp4",
				UserID:       "jerry_g",
				CreatedAt:    time.Now(),
			}

			s3URL, err := storage.UploadToS3(header.Filename, file)
			if err != nil {
				http.Error(w, "Failed to upload video", http.StatusInternalServerError)
				return
			}
			job.SourcePath = s3URL

			// Save to Database
			_, err = db.Exec(
				"INSERT INTO videos (id, user_id, status, source_path, thumbnail_url, title, description, playlist, created_at, views) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				job.ID, job.UserID, "PENDING", job.SourcePath, "", title, description, playlist, job.CreatedAt, 0,
			)

			pubsub.PublishJSON(ch, routing.ExchangeVideoTopic, routing.VideoUploadKey, job)

			// Return a 200 OK so the JavaScript knows the upload finished
			w.WriteHeader(http.StatusOK)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/status/", func(w http.ResponseWriter, r *http.Request) {
		id := filepath.Base(r.URL.Path)

		var status string
		err := db.QueryRow("SELECT status FROM videos WHERE id = ?", id).Scan(&status)
		if err != nil {
			http.Error(w, "Video ID not found in database", http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "Video ID: %s\nStatus: %s", id, status)
	})

	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("./data"))))

	http.HandleFunc("/gallery", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, status, title, playlist, source_path, thumbnail_url, views FROM videos ORDER BY created_at DESC")

		if err != nil {
			log.Printf("Query error: %v", err)
			http.Error(w, "Failed to query videos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var videos []VideoData
		for rows.Next() {
			var v VideoData
			var thumb sql.NullString
			var playlist sql.NullString
			err := rows.Scan(&v.ID, &v.Status, &v.Title, &playlist, &v.SourcePath, &thumb, &v.Views)
			if err != nil {
				continue
			}

			// Check if the db actually had a string
			if playlist.Valid {
				v.Playlist = playlist.String
			} else {
				v.Playlist = ""
			}
			if thumb.Valid && thumb.String != "" {
				v.ThumbnailURL = thumb.String
			} else {
				// Fallback to the expected S3 path while processing
				v.ThumbnailURL = fmt.Sprintf("https://%s.s3.us-east-2.amazonaws.com/%s_thumb.jpg", os.Getenv("S3_BUCKET_NAME"), v.ID)
			}

			videos = append(videos, v)
		}

		// Load and execute the template file
		tmpl, err := template.ParseFiles("web/templates/gallery.html")
		if err != nil {
			log.Printf("Template error: %v", err)
			http.Error(w, "Gallery template not found", http.StatusInternalServerError)
			return
		}

		data := GalleryPageData{Videos: videos}
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/view/", func(w http.ResponseWriter, r *http.Request) {
		id := filepath.Base(r.URL.Path)

		// 1. Check if "?embed=true" is in the URL
		isEmbed := r.URL.Query().Get("embed") == "true"

		// 2. Increment views (only if NOT an embed, or keep it to track both)
		db.Exec("UPDATE videos SET views = views + 1 WHERE id = ?", id)

		var v VideoData
		query := `SELECT id, title, description, playlist, source_path, thumbnail_url, views, created_at 
				FROM videos WHERE id = ?`

		err := db.QueryRow(query, id).Scan(
			&v.ID, &v.Title, &v.Description, &v.Playlist, &v.SourcePath, &v.ThumbnailURL, &v.Views, &v.CreatedAt,
		)

		if err != nil {
			http.Redirect(w, r, "/gallery", http.StatusSeeOther)
			return
		}

		tmpl, err := template.ParseFiles("web/templates/view.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 3. Create a map to pass both the Video and the Embed flag
		data := map[string]interface{}{
			"Video":   v,
			"IsEmbed": isEmbed,
		}

		tmpl.Execute(w, data)
	})

	http.HandleFunc("/delete/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := filepath.Base(r.URL.Path)

		storage.DeleteFromS3(id + "_processed.mp4")
		storage.DeleteFromS3(id + "_thumb.jpg")

		_, err := db.Exec("DELETE FROM videos WHERE id = ?", id)
		if err != nil {
			log.Printf("DB Delete Error: %v", err)
		}

		log.Printf("Successfully deleted video %s from S3 and DB", id)
		http.Redirect(w, r, "/gallery", http.StatusSeeOther)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/gallery", http.StatusSeeOther)
	})

	http.HandleFunc("/edit/", func(w http.ResponseWriter, r *http.Request) {
		// Extracts the ID from the URL path (e.g., /edit/vid-123 -> vid-123)
		id := filepath.Base(r.URL.Path)

		if r.Method == http.MethodPost {
			// 1. Try to get title from Form (Standard) or Query String (AJAX)
			newTitle := r.FormValue("title")
			if newTitle == "" {
				newTitle = r.URL.Query().Get("title")
			}

			// 2. Update only the title in the database
			_, err := db.Exec("UPDATE videos SET title = ? WHERE id = ?", newTitle, id)
			if err != nil {
				log.Printf("Inline edit error: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// 3. Respond with 200 OK so the JavaScript knows the save was successful
			w.WriteHeader(http.StatusOK)
			return
		}

		// If someone tries to GET this page manually, just send them back to the library
		http.Redirect(w, r, "/gallery", http.StatusSeeOther)
	})

	http.HandleFunc("/manage-thumb/", func(w http.ResponseWriter, r *http.Request) {
		id := filepath.Base(r.URL.Path)
		if r.Method == http.MethodPost {
			action := r.FormValue("thumb_action")
			var finalThumbURL string

			switch action {
			case "change":
				file, header, err := r.FormFile("new_thumbnail")
				if err == nil {
					defer file.Close()
					thumbName := fmt.Sprintf("%s_custom_%d%s", id, time.Now().Unix(), filepath.Ext(header.Filename))
					finalThumbURL, _ = storage.UploadToS3(thumbName, file)
				}
			case "remove":
				finalThumbURL = ""
			default:
				db.QueryRow("SELECT thumbnail_url FROM videos WHERE id = ?", id).Scan(&finalThumbURL)
			}

			db.Exec("UPDATE videos SET thumbnail_url = ? WHERE id = ?", finalThumbURL, id)
		}
		http.Redirect(w, r, "/gallery", http.StatusSeeOther)
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	fmt.Println("Vidify web server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
