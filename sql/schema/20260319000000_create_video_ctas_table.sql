-- +goose Up
CREATE TABLE video_ctas (
    id TEXT PRIMARY KEY,
    video_id TEXT NOT NULL,
    cta_text TEXT NOT NULL,
    cta_url TEXT NOT NULL,
    cta_time_seconds INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_video_ctas_video_id ON video_ctas(video_id);


-- +goose Down
DROP TABLE video_ctas;