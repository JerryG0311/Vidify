-- +goose Up
ALTER TABLE video_ctas ADD COLUMN impression_count INTEGER DEFAULT 0;

-- +goose Down
-- (no-op for SQLite)