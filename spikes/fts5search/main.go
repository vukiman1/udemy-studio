package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

const createIndexSQL = `CREATE VIRTUAL TABLE cue_index USING fts5(
	lecture_key UNINDEXED, lang UNINDEXED, start_ms UNINDEXED, text,
	tokenize = 'unicode61 remove_diacritics 2')`

type cue struct {
	lectureKey string
	startMs    int
	text       string
}

type searchCase struct {
	query    string
	wantHits int
}

var vietnameseStrokeFolder = strings.NewReplacer("đ", "d", "Đ", "D")

var cues = []cue{
	{"backend/001", 1000, "Mình sẽ mở một kết nối TCP tới server."},
	{"backend/002", 2000, "Đầu tiên, client gửi request và chờ response."},
	{"backend/003", 3000, "Độ trễ (latency) của handshake TLS rất quan trọng."},
	{"backend/004", 4000, "Chúng ta không dùng long polling ở đây."},
}

var searchCases = []searchCase{
	{"ket noi", 1},
	{"kết nối", 1},
	{"KET NOI", 1},
	{"tre", 1},
	{"handshake", 1},
	{"dau tien", 1},
	{"Đầu tiên", 1},
	{"độ trễ", 1},
	{"do tre", 1},
	{"chung ta", 1},
	{"ket*", 1},
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "FAIL:", err)
		os.Exit(1)
	}
}

func run() error {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	defer database.Close()

	var version string
	if err := database.QueryRow(`SELECT sqlite_version()`).Scan(&version); err != nil {
		return fmt.Errorf("read sqlite version: %w", err)
	}
	fmt.Println("sqlite_version:", version)

	if _, err := database.Exec(createIndexSQL); err != nil {
		return fmt.Errorf("create fts5 table: %w", err)
	}
	for _, item := range cues {
		if _, err := database.Exec(`INSERT INTO cue_index (lecture_key, lang, start_ms, text) VALUES (?, 'vi', ?, ?)`,
			item.lectureKey, item.startMs, vietnameseStrokeFolder.Replace(item.text)); err != nil {
			return fmt.Errorf("insert cue %s: %w", item.lectureKey, err)
		}
	}

	failures := 0
	for _, testCase := range searchCases {
		hits, err := search(database, testCase.query)
		if err != nil {
			return err
		}
		verdict := "ok"
		if len(hits) != testCase.wantHits {
			verdict = "MISMATCH"
			failures++
		}
		fmt.Printf("%-8s %-12q hits=%d want=%d %v\n", verdict, testCase.query, len(hits), testCase.wantHits, hits)
	}
	if failures > 0 {
		return fmt.Errorf("%d/%d queries mismatched", failures, len(searchCases))
	}
	fmt.Println("PASS")
	return nil
}

func search(database *sql.DB, query string) ([]string, error) {
	rows, err := database.Query(`SELECT lecture_key FROM cue_index WHERE cue_index MATCH ? ORDER BY rank`,
		vietnameseStrokeFolder.Replace(query))
	if err != nil {
		return nil, fmt.Errorf("match %q: %w", query, err)
	}
	defer rows.Close()
	var hits []string
	for rows.Next() {
		var lectureKey string
		if err := rows.Scan(&lectureKey); err != nil {
			return nil, fmt.Errorf("scan hit for %q: %w", query, err)
		}
		hits = append(hits, lectureKey)
	}
	return hits, rows.Err()
}
