-- +goose Up
-- +goose StatementBegin
create table if not exists shortify_urls (
    id bigserial primary key,
    short_url text not null unique,
    original_url text not null,
    created_at timestamp not null default now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists shortify_urls;
-- +goose StatementEnd
