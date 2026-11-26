package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite/gen"
	"github.com/vukiman1/udemy-studio/internal/transcode"
)

type CompressionSavingStore struct {
	queries *gen.Queries
}

func NewCompressionSavingStore(database *sql.DB) *CompressionSavingStore {
	return &CompressionSavingStore{queries: gen.New(database)}
}

func (store *CompressionSavingStore) Save(ctx context.Context, saving transcode.Saving) error {
	err := store.queries.SaveCompressionSaving(ctx, gen.SaveCompressionSavingParams{
		VideoPath:   saving.Path,
		CourseDir:   saving.CourseDir,
		BytesBefore: saving.BytesBefore,
		BytesAfter:  saving.BytesAfter,
	})
	if err != nil {
		return fmt.Errorf("save compression saving %q: %w", saving.Path, err)
	}
	return nil
}

func (store *CompressionSavingStore) CourseTotals(ctx context.Context, courseDir string) (transcode.SizeTotals, error) {
	row, err := store.queries.SumCourseCompressionSavings(ctx, courseDir)
	if err != nil {
		return transcode.SizeTotals{}, fmt.Errorf("sum compression savings of %q: %w", courseDir, err)
	}
	return transcode.SizeTotals{BytesBefore: row.BytesBefore, BytesAfter: row.BytesAfter}, nil
}

func (store *CompressionSavingStore) AllTotals(ctx context.Context) (transcode.SizeTotals, error) {
	row, err := store.queries.SumAllCompressionSavings(ctx)
	if err != nil {
		return transcode.SizeTotals{}, fmt.Errorf("sum compression savings: %w", err)
	}
	return transcode.SizeTotals{BytesBefore: row.BytesBefore, BytesAfter: row.BytesAfter}, nil
}
