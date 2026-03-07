# Vidify
Vidify is a high-performance, distributed video management platform that enables users to upload, transcode, and organize video content at scale. Built with a microservices-inspired architecture, it offloads heavy processing to dedicated workers to ensure a seamless and responsive user experience.

## Features
- **Distributed Processing:** Utilizes a Producer/Consumer pattern with RabbitMQ to handle asynchronous video transcoding.
- **Secure Authentication:** Multi-user support featuring Bcrypt password hashing and protected session management.
- **Cloud Native Storage:** Full integration with Amazon S3 for reliable, scalable video and thumbnail hosting.
- **Dynamic Organization:** User-defined playlist system with real-time UI updates and category badges.
- **Modern Dashboard:** - Drag-and-drop AJAX upload interface with live progress tracking.
  - Interactive gallery with Grid/List layout toggles.
  - Inline metadata editing and custom thumbnail management.

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