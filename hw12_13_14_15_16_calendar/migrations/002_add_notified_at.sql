-- +goose Up
ALTER TABLE events ADD COLUMN IF NOT EXISTS notified_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE events DROP COLUMN IF EXISTS notified_at;
