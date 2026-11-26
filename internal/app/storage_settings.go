package app

import (
	"database/sql"
	"log/slog"

	"github.com/vukiman1/udemy-studio/internal/adapter/s3bucket"
	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite"
	"github.com/vukiman1/udemy-studio/internal/objectstore"
	"github.com/vukiman1/udemy-studio/internal/platform/sealedbox"
)

func newStorageSettings(box *sealedbox.Box, database *sql.DB, logger *slog.Logger) *objectstore.Service {
	return objectstore.NewService(sqlite.NewStorageConfigStore(database, box), s3bucket.Open, logger)
}
