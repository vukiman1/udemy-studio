package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite/gen"
	"github.com/vukiman1/udemy-studio/internal/transcode"
)

type CompressionSkipStore struct {
	queries *gen.Queries
}

func NewCompressionSkipStore(database *sql.DB) *CompressionSkipStore {
	return &CompressionSkipStore{queries: gen.New(database)}
}

func (store *CompressionSkipStore) Find(ctx context.Context, path string) (transcode.Skip, bool, error) {
	row, err := store.queries.FindCompressionSkip(ctx, path)
	if errors.Is(err, sql.ErrNoRows) {
		return transcode.Skip{}, false, nil
	}
	if err != nil {
		return transcode.Skip{}, false, fmt.Errorf("find compression skip %q: %w", path, err)
	}
	return transcode.Skip{Path: row.VideoPath, SizeBytes: row.SizeBytes, Reason: row.Reason}, true, nil
}

func (store *CompressionSkipStore) Save(ctx context.Context, skip transcode.Skip) error {
	err := store.queries.SaveCompressionSkip(ctx, gen.SaveCompressionSkipParams{VideoPath: skip.Path, SizeBytes: skip.SizeBytes, Reason: skip.Reason})
	if err != nil {
		return fmt.Errorf("save compression skip %q: %w", skip.Path, err)
	}
	return nil
}
