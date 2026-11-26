-- name: SaveCompressionSaving :exec
INSERT INTO compression_savings (video_path, course_dir, bytes_before, bytes_after) VALUES (?, ?, ?, ?)
ON CONFLICT (video_path) DO UPDATE SET bytes_before = excluded.bytes_before, bytes_after = excluded.bytes_after;

-- name: SumCourseCompressionSavings :one
SELECT CAST(COALESCE(SUM(bytes_before), 0) AS INTEGER) AS bytes_before, CAST(COALESCE(SUM(bytes_after), 0) AS INTEGER) AS bytes_after
FROM compression_savings WHERE course_dir = ?;

-- name: SumAllCompressionSavings :one
SELECT CAST(COALESCE(SUM(bytes_before), 0) AS INTEGER) AS bytes_before, CAST(COALESCE(SUM(bytes_after), 0) AS INTEGER) AS bytes_after
FROM compression_savings;
