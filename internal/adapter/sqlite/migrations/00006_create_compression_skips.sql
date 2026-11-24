-- +goose Up
CREATE TABLE compression_skips (
    video_path TEXT PRIMARY KEY,
    size_bytes INTEGER NOT NULL,
    reason     TEXT NOT NULL
) STRICT;

-- +goose Down
DROP TABLE compression_skips;
