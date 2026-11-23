-- name: UpsertJob :exec
INSERT INTO jobs (id, status, created_at, state_json) VALUES (?, ?, ?, ?)
ON CONFLICT (id) DO UPDATE SET status = excluded.status, state_json = excluded.state_json;

-- name: ListJobs :many
SELECT state_json FROM jobs ORDER BY created_at DESC;
