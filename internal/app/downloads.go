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
	"github.com/vukiman1/udemy-studio/internal/progress"
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

func newDownloads(ctx context.Context, config DownloadOptions, dataDir string, database *sql.DB, catalog *library.Catalog, progressService *progress.Service, logger *slog.Logger) (downloads, error) {
	if !config.isConfigured() {
		logger.Info("downloads disabled; pass --python and --python-repo to enable them")
		return downloads{}, nil
	}
	loginService := newLoginService(config.ChromeDataDir, database)
	downloader := pythondownloader.New(pythondownloader.Config{
		Python:          config.Python,
		RepoDir:         config.PythonRepo,
		OutputDir:       config.OutputDir,
		LogDir:          filepath.Join(dataDir, jobLogsDirName),
		CaptionLanguage: config.CaptionLanguage,
		KeepAwake:       true,
	}, logger)
	manager := jobs.NewManager(jobs.Deps{
		Downloader: downloader,
		Tokens:     loginService,
		Store:      sqlite.NewJobStore(database),
		Logger:     logger,
		Now:        time.Now,
		OnSucceeded: func(ctx context.Context, job jobs.Job) {
			if err := catalog.Rescan(); err != nil {
				logger.Error("rescan library after download", "job", job.ID, "err", err)
				return
			}
			if err := progressService.AdoptKeys(ctx, catalog.Current().KeyRenames()); err != nil {
				logger.Error("move progress to the lecture ids of the new download", "job", job.ID, "err", err)
			}
		},
	})
	if err := manager.Restore(ctx); err != nil {
		return downloads{}, fmt.Errorf("restore download jobs: %w", err)
	}
	logger.Info("downloads enabled", "python", config.Python, "output", config.OutputDir)
	return downloads{login: loginService, jobs: manager}, nil
}

func (configured downloads) workers() []func(context.Context) error {
	if configured.jobs == nil {
		return nil
	}
	return []func(context.Context) error{configured.jobs.Run}
}
