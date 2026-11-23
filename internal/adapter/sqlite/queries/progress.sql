-- name: FindProgress :one
SELECT lecture_key, position_seconds, duration_seconds, completed_at, updated_at
FROM lecture_progress
WHERE lecture_key = ?;

-- name: UpsertProgress :exec
INSERT INTO lecture_progress (lecture_key, position_seconds, duration_seconds, completed_at, updated_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT (lecture_key) DO UPDATE SET
    position_seconds = excluded.position_seconds,
    duration_seconds = excluded.duration_seconds,
    completed_at     = excluded.completed_at,
    updated_at       = excluded.updated_at;

-- name: ListProgress :many
SELECT lecture_key, position_seconds, duration_seconds, completed_at, updated_at
FROM lecture_progress
ORDER BY updated_at DESC;
