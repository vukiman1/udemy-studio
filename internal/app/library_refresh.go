package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/progress"
	"github.com/vukiman1/udemy-studio/internal/search"
)

// libraryRefresher keeps what is derived from the library, progress keys and the subtitle index, in step with it.
type libraryRefresher struct {
	catalog  *library.Catalog
	media    fs.FS
	progress *progress.Service
	search   *search.Service
	logger   *slog.Logger
}

func (refresher libraryRefresher) afterDownload(ctx context.Context) error {
	if err := refresher.catalog.Rescan(); err != nil {
		return fmt.Errorf("rescan library: %w", err)
	}
	return errors.Join(
		refresher.progress.AdoptKeys(ctx, refresher.catalog.Current().KeyRenames()),
		refresher.search.Rebuild(ctx, refresher.catalog, refresher.media),
	)
}

// indexSubtitles runs in the background at start, so the player answers while the index is built.
func (refresher libraryRefresher) indexSubtitles(ctx context.Context) error {
	started := time.Now()
	// A failed rebuild leaves the last index in place, so it must not stop the server.
	switch err := refresher.search.Rebuild(ctx, refresher.catalog, refresher.media); {
	case err == nil:
		refresher.logger.Info("subtitles indexed", "took", time.Since(started).Round(time.Millisecond))
	case ctx.Err() == nil:
		refresher.logger.Error("subtitle index not refreshed; search answers from the last one", "err", err)
	}
	return nil
}
