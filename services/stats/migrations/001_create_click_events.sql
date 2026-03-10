-- +goose Up
CREATE TABLE click_events (
    code       String,
    ip         String,
    country    String,
    clicked_at DateTime
) ENGINE = MergeTree()
ORDER BY (code, clicked_at);

-- +goose Down
DROP TABLE click_events;