-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id          TEXT PRIMARY KEY,
    title       TEXT        NOT NULL,
    start_at    TIMESTAMPTZ NOT NULL,
    duration    BIGINT      NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    user_id     TEXT        NOT NULL,
    notify_at   BIGINT      NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_events_start_at ON events (start_at);

-- +goose Down
DROP TABLE IF EXISTS events;
