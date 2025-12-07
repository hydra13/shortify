-- +goose Up
-- +goose StatementBegin
ALTER TABLE shortify_urls ADD COLUMN IF NOT EXISTS user_id TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE shortify_urls DROP COLUMN IF EXISTS user_id;
-- +goose StatementEnd
