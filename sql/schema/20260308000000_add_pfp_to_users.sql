-- +goose Up
ALTER TABLE users ADD COLUMN profile_picture_url TEXT;


-- +goose Down
ALTER TABLE users DROP COLUMN profile_picture_url;