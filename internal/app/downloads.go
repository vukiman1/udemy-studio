package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite"
	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/login"
	"github.com/vukiman1/udemy-studio/internal/platform/sealedbox"
)

const jobLogsDirName = "job-logs"

type DownloadOptions struct {
	Tools           ToolPaths
	OutputDir       string
	ChromeDataDir   string
	CaptionLanguage string
}

// canFindTools logs the first tool downloads cannot run, so jobs are not offered only to fail.
func (options DownloadOptions) canFindTools(logger *slog.Logger) bool {
	required := []struct{ path, flag string }{
		{options.Tools.Aria2c, "--aria2c"},
		{options.Tools.YTDLP, "--yt-dlp"},
	}
	for _, tool := range required {
		if _, err := exec.LookPath(tool.path); err != nil {
			logger.Warn("downloads disabled: "+tool.path+" cannot run; install it or pass its path with "+tool.flag, "err", err)
			return false
		}
	}
	return true
}

type downloads struct {
	login *login.Service
	jobs  *jobs.Manager
}

func newDownloads(ctx context.Context, config DownloadOptions, dataDir string, database *sql.DB, box *sealedbox.Box, onDownloaded func(context.Context) error, logger *slog.Logger) (downloads, error) {
	if !config.canFindTools(logger) {
		return downloads{}, nil
	}
	downloader := jobDownloader{
		outputDir:       config.OutputDir,
		logDir:          filepath.Join(dataDir, jobLogsDirName),
		captionLanguage: config.CaptionLanguage,
		tools:           config.Tools,
	}
	loginService := newLoginService(config.ChromeDataDir, database, box, logger)
	manager := jobs.NewManager(jobs.Deps{
		Downloader:    downloader,
		Tokens:        loginService,
		Store:         sqlite.NewJobStore(database),
		Logger:        logger,
		Now:           time.Now,
		ClassifyError: classifyDownloadError,
		OnSucceeded: func(ctx context.Context, job jobs.Job) {
			if err := onDownloaded(ctx); err != nil {
				logger.Error("refresh the library after a download", "job", job.ID, "err", err)
			}
		},
	})
	if err := manager.Restore(ctx); err != nil {
		return downloads{}, fmt.Errorf("restore download jobs: %w", err)
	}
	logger.Info("downloads enabled", "output", config.OutputDir)
	return downloads{login: loginService, jobs: manager}, nil
}

func (configured downloads) workers() []func(context.Context) error {
	if configured.jobs == nil {
		return nil
	}
	return []func(context.Context) error{configured.jobs.Run}
}
