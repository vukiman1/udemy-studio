package app

import (
	"io/fs"
	"log/slog"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/objectstore"
	"github.com/vukiman1/udemy-studio/internal/publish"
)

func newPublish(storage *objectstore.Service, catalog *library.Catalog, files fs.FS, logger *slog.Logger) *publish.Service {
	return publish.NewService(publish.Deps{
		Catalog:    catalog,
		Files:      files,
		Buckets:    storage,
		LimitBytes: publish.DefaultLimitBytes,
		Now:        time.Now,
		Logger:     logger,
	})
}
