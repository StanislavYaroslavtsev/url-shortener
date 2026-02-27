-- +goose Up
CREATE TABLE links (
    code VARCHAR(6) PRIMARY KEY,
    url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE links;