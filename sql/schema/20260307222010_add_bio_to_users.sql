-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN bio TEXT DEFAULT 'No bio set.';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down migration not supported in sqlite for column drop';
-- +goose StatementEnd
