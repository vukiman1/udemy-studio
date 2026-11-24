-- +goose Up
CREATE TABLE notes (
    id          INTEGER PRIMARY KEY,
    lecture_key TEXT NOT NULL,
    at_seconds  REAL NOT NULL,
    body        TEXT NOT NULL,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
) STRICT;

CREATE INDEX notes_by_lecture ON notes (lecture_key, at_seconds);

-- +goose Down
DROP TABLE notes;
