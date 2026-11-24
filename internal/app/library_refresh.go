package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/notes"
	"github.com/vukiman1/udemy-studio/internal/progress"
	"github.com/vukiman1/udemy-studio/internal/search"
)

// libraryRefresher keeps what is derived from the library, lecture keys and the subtitle index, in step with it.
type libraryRefresher struct {
	catalog  *library.Catalog
	media    fs.FS
	progress *progress.Service
	notes    *notes.Service
	search   *search.Service
	logger   *slog.Logger
}

func (refresher libraryRefresher) afterDownload(ctx context.Context) error {
	if err := refresher.catalog.Rescan(); err != nil {
		return fmt.Errorf("rescan library: %w", err)
	}
	return errors.Join(
		adoptLectureKeys(ctx, refresher.catalog, refresher.progress, refresher.notes),
		refresher.search.Rebuild(ctx, refresher.catalog, refresher.media),
	)
}

// adoptLectureKeys moves what was saved under a lecture's path key once a manifest gives it its Udemy id.
func adoptLectureKeys(ctx context.Context, catalog *library.Catalog, progressService *progress.Service, notesService *notes.Service) error {
	renames := catalog.Current().KeyRenames()
	return errors.Join(progressService.AdoptKeys(ctx, renames), notesService.AdoptKeys(ctx, renames))
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
