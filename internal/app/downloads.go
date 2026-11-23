package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/vukiman1/udemy-studio/internal/adapter/pythondownloader"
	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite"
	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/login"
)

const jobLogsDirName = "job-logs"

type DownloadOptions struct {
	Python          string
	PythonRepo      string
	OutputDir       string
	ChromeDataDir   string
	CaptionLanguage string
}

func (options DownloadOptions) isConfigured() bool {
	return options.Python != "" && options.PythonRepo != ""
}

type downloads struct {
	login *login.Service
	jobs  *jobs.Manager
}

func newDownloads(ctx context.Context, options Options, database *sql.DB, catalog *library.Catalog, logger *slog.Logger) (downloads, error) {
	config := options.Downloads
	if !config.isConfigured() {
		logger.Info("downloads disabled; pass --python and --python-repo to enable them")
		return downloads{}, nil
	}
	loginService := newLoginService(config.ChromeDataDir, database)
	downloader := pythondownloader.New(pythondownloader.Config{
		Python:          config.Python,
		RepoDir:         config.PythonRepo,
		OutputDir:       config.OutputDir,
		LogDir:          filepath.Join(filepath.Dir(options.DatabasePath), jobLogsDirName),
		CaptionLanguage: config.CaptionLanguage,
		KeepAwake:       true,
	}, logger)
	manager := jobs.NewManager(jobs.Deps{
		Downloader: downloader,
		Tokens:     loginService,
		Store:      sqlite.NewJobStore(database),
		Logger:     logger,
		Now:        time.Now,
		OnSucceeded: func(_ context.Context, job jobs.Job) {
			if err := catalog.Rescan(); err != nil {
				logger.Error("rescan library after download", "job", job.ID, "err", err)
			}
		},
	})
	if err := manager.Restore(ctx); err != nil {
		return downloads{}, fmt.Errorf("restore download jobs: %w", err)
	}
	logger.Info("downloads enabled", "python", config.Python, "output", config.OutputDir)
	return downloads{login: loginService, jobs: manager}, nil
}

func (enabled downloads) workers() []func(context.Context) error {
	if enabled.jobs == nil {
		return nil
	}
	return []func(context.Context) error{enabled.jobs.Run}
}
