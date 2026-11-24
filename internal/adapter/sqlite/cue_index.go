package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vukiman1/udemy-studio/internal/search"
)

// sqlc cannot parse FTS5's MATCH on the table name, so these statements are written by hand.
const (
	clearCuesSQL  = `DELETE FROM cue_index`
	insertCueSQL  = `INSERT INTO cue_index (lecture_key, lang, start_ms, text, search_text) VALUES (?, ?, ?, ?, ?)`
	searchCuesSQL = `SELECT lecture_key, lang, start_ms, text, search_text, highlight(cue_index, 4, ?, ?)
FROM cue_index WHERE cue_index MATCH ? ORDER BY rank LIMIT ?`
	prefixOperator = "*"
)

type CueIndex struct {
	database *sql.DB
}

func NewCueIndex(database *sql.DB) *CueIndex {
	return &CueIndex{database: database}
}

// Replace swaps the whole index in one transaction, so a search never sees it half built.
func (index *CueIndex) Replace(ctx context.Context, entries []search.Entry) error {
	transaction, err := index.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin cue index rebuild: %w", err)
	}
	if err := insertCues(ctx, transaction, entries); err != nil {
		return errors.Join(err, transaction.Rollback())
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit cue index rebuild: %w", err)
	}
	return nil
}

func insertCues(ctx context.Context, transaction *sql.Tx, entries []search.Entry) error {
	if _, err := transaction.ExecContext(ctx, clearCuesSQL); err != nil {
		return fmt.Errorf("clear cue index: %w", err)
	}
	insert, err := transaction.PrepareContext(ctx, insertCueSQL)
	if err != nil {
		return fmt.Errorf("prepare cue insert: %w", err)
	}
	defer insert.Close()
	for _, entry := range entries {
		if _, err := insert.ExecContext(ctx, entry.LectureKey, entry.Language, entry.Start.Milliseconds(), entry.Text, entry.SearchText); err != nil {
			return fmt.Errorf("index a cue of %q: %w", entry.LectureKey, err)
		}
	}
	return nil
}

func (index *CueIndex) Search(ctx context.Context, terms []string, limit int) ([]search.Match, error) {
	rows, err := index.database.QueryContext(ctx, searchCuesSQL,
		search.HighlightStart, search.HighlightEnd, matchExpression(terms), limit)
	if err != nil {
		return nil, fmt.Errorf("query cue index: %w", err)
	}
	defer rows.Close()
	var matches []search.Match
	for rows.Next() {
		var match search.Match
		var startMs int64
		if err := rows.Scan(&match.LectureKey, &match.Language, &startMs, &match.Text, &match.SearchText, &match.HighlightedSearchText); err != nil {
			return nil, fmt.Errorf("read a cue match: %w", err)
		}
		match.Start = time.Duration(startMs) * time.Millisecond
		matches = append(matches, match)
	}
	return matches, rows.Err()
}

// matchExpression quotes every term, so FTS5 syntax typed by the user is searched for rather than interpreted.
func matchExpression(terms []string) string {
	quoted := make([]string, len(terms))
	for position, term := range terms {
		quoted[position] = `"` + strings.ReplaceAll(term, `"`, `""`) + `"`
	}
	quoted[len(quoted)-1] += prefixOperator
	return strings.Join(quoted, " ")
}
