-- +goose Up
CREATE TABLE jobs (
    id         TEXT PRIMARY KEY,
    status     TEXT NOT NULL,
    created_at TEXT NOT NULL,
    state_json TEXT NOT NULL
) STRICT;

-- +goose Down
DROP TABLE jobs;
