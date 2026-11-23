package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	jobLogDirMode  = 0o750
	jobLogFileMode = 0o600
	jobLogSuffix   = ".log"
)

// goDownloader runs a download job in process, writing what it and the tools print to the job's log like the Python runner.
type goDownloader struct {
	outputDir       string
	logDir          string
	captionLanguage string
	tools           ToolPaths
}

func (downloader goDownloader) Download(ctx context.Context, job jobs.Job, token secret.Secret, onEvent func(jobs.Event)) error {
	request, err := newCourseDownload(job.Request.CourseURL, job.Request.Chapters)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(downloader.logDir, jobLogDirMode); err != nil {
		return fmt.Errorf("create job log dir: %w", err)
	}
	logPath := filepath.Join(downloader.logDir, job.ID+jobLogSuffix)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, jobLogFileMode) //nolint:gosec // reason: the configured log dir plus a generated job id
	if err != nil {
		return fmt.Errorf("create job log %s: %w", logPath, err)
	}
	defer logFile.Close()
	logger := slog.New(slog.NewTextHandler(logFile, nil))
	request.outputDir, request.captionLanguage = downloader.outputDir, downloader.captionLanguage
	request.tools, request.toolOutput = downloader.tools, logFile

	client, err := visitedClient(ctx, request.ref, token, logger)
	if err == nil {
		_, err = downloadCourse(ctx, client, request, logger, onEvent)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("go downloader failed (see %s): %w", logPath, err)
	}
	return nil
}
