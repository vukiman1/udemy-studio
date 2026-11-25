package download

import (
	"context"
	"path/filepath"

	"github.com/vukiman1/udemy-studio/internal/library"
)

// cover fetches the course's cover image once. A failure costs only the picture, so it is logged rather than raised;
// only a cancel stops the download.
func (run *courseDownload) cover(ctx context.Context, courseDir string) error {
	if run.options.CoverURL == "" || isDownloaded(filepath.Join(courseDir, library.CoverFileName)) {
		return nil
	}
	err := run.tools.Files.FetchFile(ctx, run.options.CoverURL, courseDir, library.CoverFileName)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		run.logger.Warn("course cover not downloaded", "url", run.options.CoverURL, "err", err)
	}
	return nil
}
