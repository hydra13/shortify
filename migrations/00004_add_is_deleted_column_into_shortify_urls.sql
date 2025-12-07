-- +goose Up
-- +goose StatementBegin
ALTER TABLE shortify_urls ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE shortify_urls DROP COLUMN IF EXISTS is_deleted;
-- +goose StatementEnd
