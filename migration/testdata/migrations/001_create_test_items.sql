-- +goose Up
CREATE TABLE IF NOT EXISTS test_items (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS test_items;
