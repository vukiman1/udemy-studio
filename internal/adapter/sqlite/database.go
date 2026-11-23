package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"net/url"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

const (
	driverName     = "sqlite"
	migrationsDir  = "migrations"
	busyTimeoutMs  = "5000"
	maxConnections = 1
)

//go:embed migrations/*.sql
var migrations embed.FS

func Open(ctx context.Context, path string) (*sql.DB, error) {
	pragmas := url.Values{"_pragma": {"journal_mode(WAL)", "foreign_keys(1)", "busy_timeout(" + busyTimeoutMs + ")"}}
	dataSource := url.URL{Scheme: "file", Path: path, RawQuery: pragmas.Encode()}
	database, err := sql.Open(driverName, dataSource.String())
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", path, err)
	}
	database.SetMaxOpenConns(maxConnections)
	if err := migrate(ctx, database); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("migrate sqlite %s: %w", path, err)
	}
	return database, nil
}

func migrate(ctx context.Context, database *sql.DB) error {
	migrationFiles, err := fs.Sub(migrations, migrationsDir)
	if err != nil {
		return fmt.Errorf("open embedded migrations: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, database, migrationFiles)
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
