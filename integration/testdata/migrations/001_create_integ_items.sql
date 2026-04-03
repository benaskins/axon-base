-- +goose Up
CREATE TABLE IF NOT EXISTS integ_items (
    id   TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS integ_items;
