-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE,
    password TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS videos (
    id TEXT PRIMARY KEY, 
    user_id TEXT,
    status TEXT,
    source_path TEXT,
    thumbnail_url TEXT,
    title TEXT,
    description TEXT,
    playlist TEXT,
    views INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(email)
);

-- +goose Down
DROP TABLE IF EXISTS videos;
DROP TABLE IF EXISTS users;