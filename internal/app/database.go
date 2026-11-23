package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite"
)

const databaseFileName = "app.db"

// openDatabase opens the database inside dataDir, which also holds the LAN access code and download logs.
func openDatabase(ctx context.Context, dataDir string) (*sql.DB, error) {
	if err := os.MkdirAll(dataDir, dataDirMode); err != nil {
		return nil, fmt.Errorf("create data dir %s: %w", dataDir, err)
	}
	return sqlite.Open(ctx, filepath.Join(dataDir, databaseFileName))
}
