# Go binaries
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY . .
RUN go mod download

# --- NEW: Install Goose ---
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

RUN go build -o api ./cmd/api/main.go
RUN go build -o worker ./cmd/worker/main.go


# Final lightweight image
FROM alpine:latest
RUN apk add --no-cache ffmpeg ca-certificates libc6-compat
WORKDIR /app

# --- NEW: Copy Goose to the final image ---
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# 1. Copying the 'api' binary and naming it 'api' in the local dir
COPY --from=builder /app/api ./api
COPY --from=builder /app/worker ./worker

# 2. Copying the web folder from the builder page
COPY --from=builder /app/web ./web

# Expose the API port
RUN mkdir -p data
EXPOSE 8080

# Use the full path to be 100% sure
CMD ["./api"]