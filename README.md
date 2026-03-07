# Vidify
A robust, distributed video management and transcoding platform built with Go, RabbitMQ, and AWS S3.

## Features
- **Distributed Architecture:** Uses a Worker/API pattern to handle heavy video transcoding without blocking the web server.
- **User Authentication:** Secure Signup/Login system using Bcrypt password hashing and session-based cookies.
- **Async Processing:** RabbitMQ manages the job queue, ensuring reliable video processing.
- **Cloud Storage:** Integrated with Amazon S3 for durable video and thumbnail storage.
- **Modern Dashboard:**
  - Drag-and-drop AJAX uploads with real-time progress bars.
  - Playlist organization system with dynamic badges.
  - Inline title editing and custom thumbnail management.
  - Grid and List view toggles.

## Tech Stack
- **Backend:** Go (Golang)
- **Database:** SQLite3
- **Messaging:** RabbitMQ
- **Storage:** AWS S3
- **DevOps:** Docker and Docker Compose
- **Frontend:** HTML5, CSS3 (Inter font), Vanilla JavaScript

## Getting Started
1. Clone the repository.
2. Set up your environment variables in docker-compose.yml (AWS Credentials, S3 Bucket Name).
3. Run the application:
   ```bash
   docker-compose up --build
4. Access the app at http://localhost:8080.
