-- name: CreateNote :one
INSERT INTO notes (lecture_key, at_seconds, body, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
RETURNING id, lecture_key, at_seconds, body, created_at, updated_at;

-- name: UpdateNoteBody :one
UPDATE notes SET body = ?, updated_at = ? WHERE id = ?
RETURNING id, lecture_key, at_seconds, body, created_at, updated_at;

-- name: DeleteNote :execrows
DELETE FROM notes WHERE id = ?;

-- name: ListLectureNotes :many
SELECT id, lecture_key, at_seconds, body, created_at, updated_at
FROM notes
WHERE lecture_key = ?
ORDER BY at_seconds, id;

-- name: ListNotes :many
SELECT id, lecture_key, at_seconds, body, created_at, updated_at
FROM notes
ORDER BY lecture_key, at_seconds, id;

-- name: RenameNotesKey :exec
UPDATE notes SET lecture_key = sqlc.arg(to_key) WHERE lecture_key = sqlc.arg(from_key);
