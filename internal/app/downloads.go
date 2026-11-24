package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/vukiman1/udemy-studio/internal/adapter/pythondownloader"
	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite"
	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/login"
)

const (
	jobLogsDirName   = "job-logs"
	DownloaderPython = "python"
	DownloaderGo     = "go"
)

var ErrUnknownDownloader = errors.New("unknown downloader: use " + DownloaderPython + " or " + DownloaderGo)

type DownloadOptions struct {
	Downloader      string
	Python          string
	PythonRepo      string
	Tools           ToolPaths
	OutputDir       string
	ChromeDataDir   string
	CaptionLanguage string
}

// isConfigured is false only for the Python downloader without its checkout; any other value reaches newDownloader.
func (options DownloadOptions) isConfigured() bool {
	return options.Downloader != DownloaderPython || (options.Python != "" && options.PythonRepo != "")
}

// missingTool names the first tool the Go downloader cannot find, so jobs are not offered only to fail.
func (options DownloadOptions) missingTool() (string, bool) {
	if options.Downloader != DownloaderGo {
		return "", false
	}
	for _, tool := range []string{options.Tools.Aria2c, options.Tools.YTDLP} {
		if _, err := exec.LookPath(tool); err != nil {
			return filepath.Base(tool), true
		}
	}
	return "", false
}

func (options DownloadOptions) newDownloader(logDir string, logger *slog.Logger) (jobs.Downloader, error) {
	switch options.Downloader {
	case DownloaderGo:
		return goDownloader{outputDir: options.OutputDir, logDir: logDir, captionLanguage: options.CaptionLanguage, tools: options.Tools}, nil
	case DownloaderPython:
		return pythondownloader.New(pythondownloader.Config{
			Python:          options.Python,
			RepoDir:         options.PythonRepo,
			OutputDir:       options.OutputDir,
			LogDir:          logDir,
			CaptionLanguage: options.CaptionLanguage,
			KeepAwake:       true,
		}, logger), nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownDownloader, options.Downloader)
	}
}

type downloads struct {
	login *login.Service
	jobs  *jobs.Manager
}

func newDownloads(ctx context.Context, config DownloadOptions, dataDir string, database *sql.DB, onDownloaded func(context.Context) error, logger *slog.Logger) (downloads, error) {
	if !config.isConfigured() {
		logger.Info("downloads disabled; the Python downloader needs --python and --python-repo, or use --downloader go")
		return downloads{}, nil
	}
	if tool, missing := config.missingTool(); missing {
		logger.Warn("downloads disabled: the Go downloader needs "+tool+"; install it or pass its path with --"+tool, "downloader", config.Downloader)
		return downloads{}, nil
	}
	downloader, err := config.newDownloader(filepath.Join(dataDir, jobLogsDirName), logger)
	if err != nil {
		return downloads{}, err
	}
	loginService := newLoginService(config.ChromeDataDir, database)
	manager := jobs.NewManager(jobs.Deps{
		Downloader: downloader,
		Tokens:     loginService,
		Store:      sqlite.NewJobStore(database),
		Logger:     logger,
		Now:        time.Now,
		OnSucceeded: func(ctx context.Context, job jobs.Job) {
			if err := onDownloaded(ctx); err != nil {
				logger.Error("refresh the library after a download", "job", job.ID, "err", err)
			}
		},
	})
	if err := manager.Restore(ctx); err != nil {
		return downloads{}, fmt.Errorf("restore download jobs: %w", err)
	}
	logger.Info("downloads enabled", "downloader", config.Downloader, "output", config.OutputDir)
	return downloads{login: loginService, jobs: manager}, nil
}

func (configured downloads) workers() []func(context.Context) error {
	if configured.jobs == nil {
		return nil
	}
	return []func(context.Context) error{configured.jobs.Run}
}
