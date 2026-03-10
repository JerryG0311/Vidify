-- +goose Up
ALTER TABLE users ADD COLUMN display_name TEXT;
ALTER TABLE users ADD COLUMN username TEXT;
ALTER TABLE users ADD COLUMN website TEXT;
ALTER TABLE users ADD COLUMN instagram TEXT;
CREATE UNIQUE INDEX idx_users_username ON users(username);

-- +goose Down
DROP INDEX IF EXISTS idx_users_username;
ALTER TABLE users DROP COLUMN display_name;
ALTER TABLE users DROP COLUMN username;
ALTER TABLE users DROP COLUMN website;
ALTER TABLE users DROP COLUMN instagram;