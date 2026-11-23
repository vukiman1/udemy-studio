package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite/gen"
	"github.com/vukiman1/udemy-studio/internal/progress"
)

type ProgressStore struct {
	queries *gen.Queries
}

func NewProgressStore(database *sql.DB) *ProgressStore {
	return &ProgressStore{queries: gen.New(database)}
}

func (store *ProgressStore) Find(ctx context.Context, lectureKey string) (progress.Record, bool, error) {
	row, err := store.queries.FindProgress(ctx, lectureKey)
	if errors.Is(err, sql.ErrNoRows) {
		return progress.Record{}, false, nil
	}
	if err != nil {
		return progress.Record{}, false, fmt.Errorf("find progress %q: %w", lectureKey, err)
	}
	record, err := toRecord(row)
	if err != nil {
		return progress.Record{}, false, err
	}
	return record, true, nil
}

func (store *ProgressStore) Upsert(ctx context.Context, record progress.Record) error {
	params := gen.UpsertProgressParams{
		LectureKey:      record.LectureKey,
		PositionSeconds: record.PositionSeconds,
		DurationSeconds: record.DurationSeconds,
		UpdatedAt:       formatTime(record.UpdatedAt),
	}
	if record.CompletedAt != nil {
		params.CompletedAt = sql.NullString{String: formatTime(*record.CompletedAt), Valid: true}
	}
	if err := store.queries.UpsertProgress(ctx, params); err != nil {
		return fmt.Errorf("upsert progress %q: %w", record.LectureKey, err)
	}
	return nil
}

func (store *ProgressStore) RenameKey(ctx context.Context, from, to string) error {
	if err := store.queries.RenameProgressKey(ctx, gen.RenameProgressKeyParams{FromKey: from, ToKey: to}); err != nil {
		return fmt.Errorf("rename progress %q to %q: %w", from, to, err)
	}
	return nil
}

func (store *ProgressStore) All(ctx context.Context) ([]progress.Record, error) {
	rows, err := store.queries.ListProgress(ctx)
	if err != nil {
		return nil, fmt.Errorf("list progress: %w", err)
	}
	records := make([]progress.Record, 0, len(rows))
	for _, row := range rows {
		record, err := toRecord(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func toRecord(row gen.LectureProgress) (progress.Record, error) {
	updatedAt, err := parseTime(row.UpdatedAt)
	if err != nil {
		return progress.Record{}, fmt.Errorf("progress %q updated_at: %w", row.LectureKey, err)
	}
	record := progress.Record{
		LectureKey:      row.LectureKey,
		PositionSeconds: row.PositionSeconds,
		DurationSeconds: row.DurationSeconds,
		UpdatedAt:       updatedAt,
	}
	if row.CompletedAt.Valid {
		completedAt, err := parseTime(row.CompletedAt.String)
		if err != nil {
			return progress.Record{}, fmt.Errorf("progress %q completed_at: %w", row.LectureKey, err)
		}
		record.CompletedAt = &completedAt
	}
	return record, nil
}
