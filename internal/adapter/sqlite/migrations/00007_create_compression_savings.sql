-- +goose Up
CREATE TABLE compression_savings (
    video_path   TEXT PRIMARY KEY,
    course_dir   TEXT NOT NULL,
    bytes_before INTEGER NOT NULL,
    bytes_after  INTEGER NOT NULL
) STRICT;

-- +goose Down
DROP TABLE compression_savings;
