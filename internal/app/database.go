package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite"
)

func openDatabase(ctx context.Context, path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), dataDirMode); err != nil {
		return nil, fmt.Errorf("create data dir for %s: %w", path, err)
	}
	return sqlite.Open(ctx, path)
}
