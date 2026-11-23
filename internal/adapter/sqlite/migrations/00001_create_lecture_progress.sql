-- +goose Up
CREATE TABLE lecture_progress (
    lecture_key      TEXT PRIMARY KEY,
    position_seconds REAL NOT NULL,
    duration_seconds REAL NOT NULL,
    completed_at     TEXT,
    updated_at       TEXT NOT NULL
) STRICT;

-- +goose Down
DROP TABLE lecture_progress;
