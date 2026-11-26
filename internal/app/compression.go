package app

import (
	"database/sql"
	"log/slog"
	"os"
	"os/exec"

	"github.com/vukiman1/udemy-studio/internal/adapter/ffmpeg"
	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite"
	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/transcode"
)

type CompressionOptions struct {
	FFmpeg  string
	FFprobe string
}

// newCompression returns nil when a tool is missing, so the page offers no button that could only fail.
func newCompression(options CompressionOptions, libraryDir string, media *os.Root, database *sql.DB, catalog *library.Catalog, logger *slog.Logger) *transcode.Service {
	required := []struct{ binary, hint string }{
		{options.FFmpeg, "install it or pass its path with --ffmpeg"},
		{options.FFprobe, "install it or pass its path with --ffprobe"},
		{ffmpeg.NiceBinary, "encodes run under it so they yield to the learner"},
	}
	for _, tool := range required {
		if _, err := exec.LookPath(tool.binary); err != nil {
			logger.Warn("compression disabled: "+tool.binary+" not found; "+tool.hint, "err", err)
			return nil
		}
	}
	tools := ffmpeg.Tools{FFmpeg: options.FFmpeg, FFprobe: options.FFprobe, LibraryDir: libraryDir}
	return transcode.NewService(transcode.Deps{
		Catalog: catalog,
		Prober:  tools,
		Encoder: tools,
		Files:   media,
		Skips:   sqlite.NewCompressionSkipStore(database),
		Savings: sqlite.NewCompressionSavingStore(database),
		Logger:  logger,
	})
}
