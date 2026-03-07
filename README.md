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

## Motivation
Managing video content is computationally expensive and traditionally slows down web applications. I built Vidify because I wanted to solve the "blocking problem" in standard web architectures. By decoupling the video upload and transcoding processes using a distributed worker system and RabbitMQ, I created a platform that remains snappy and responsive for the user, regardless of the processing load on the backend. This project allowed me to dive deep into asynchronous systems, cloud storage durability, and multi-user authentication.

## Usage

Vidify is designed to be intuitive. Users can manage their video library through a modern web interface that communicates with a distributed backend.

### User Workflow
1. **Account Management:** Users can create an account via `/signup` and securely authenticate via `/login`.
2. **Uploading Content:** Use the drag-and-drop interface to upload videos. Users can specify a Title and a Playlist Name to categorize their content.
3. **Real-time Monitoring:** The dashboard automatically polls the backend to update the processing status (Pending -> Completed) without requiring a page refresh.
4. **Library Management:**
   - **Toggles:** Switch between Grid and List views to suit your preference.
   - **Editing:** Click on any video title to edit it inline. Use the options menu (three dots) to upload custom thumbnails or delete videos.
   - **Social Sharing:** Click the share icon to generate custom links for X, Facebook, and LinkedIn.

### Local Development Usage
To test the distributed nature of the application, you can scale the worker service to handle multiple concurrent transcoding jobs:

```bash
docker-compose up --build --scale worker=3

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


## Contributing

Contributions are welcome! If you would like to help improve Vidify, please follow these steps to get your local development environment running.

### Clone the repo

```bash
git clone https://github.com/JerryG0311/Vidify
cd Vidify
```

### Build and run for development

The project is containerized to manage the Go environment, SQLite, and RabbitMQ dependencies automatically. Ensure you have Docker installed.

```bash
# Start all services in the background
docker-compose up -d --build

# Follow the API logs to see real-time interaction
docker-compose logs -f api
```

### Run the test suite

To ensure core logic remains stable after your changes, run the test suite from the root directory:

```bash
go test ./...
```

### Submit a pull request

1. Fork the repository on GitHub.
2. Create a feature branch: `git checkout -b feature/amazing-feature`.
3. Commit your changes: `git commit -m 'feat: add amazing feature'`.
4. Push to your branch: `git push origin feature/amazing-feature`.
5. Open a Pull Request to the main branch.