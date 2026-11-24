-- name: FindCompressionSkip :one
SELECT video_path, size_bytes, reason FROM compression_skips WHERE video_path = ?;

-- name: SaveCompressionSkip :exec
INSERT INTO compression_skips (video_path, size_bytes, reason) VALUES (?, ?, ?)
ON CONFLICT (video_path) DO UPDATE SET size_bytes = excluded.size_bytes, reason = excluded.reason;
