package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/vukiman1/udemy-studio/internal/adapter/downloadtools"
	"github.com/vukiman1/udemy-studio/internal/download"
	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const playlistTempPattern = "udemy-studio-hls-*"

type ToolPaths struct {
	Aria2c string
	YTDLP  string
}

type DownloadCourseOptions struct {
	CourseURL       string
	Chapters        string
	OutputDir       string
	DataDir         string
	ChromeDataDir   string
	CaptionLanguage string
	Tools           ToolPaths
	ToolOutput      io.Writer
}

// DownloadCourse plans a course with the selected Chrome profile's login and downloads it as the Python downloader would.
func DownloadCourse(ctx context.Context, options DownloadCourseOptions, logger *slog.Logger, onEvent func(jobs.Event)) (download.Summary, error) {
	chapters, err := download.ParseChapterFilter(options.Chapters)
	if err != nil {
		return download.Summary{}, err
	}
	ref, err := udemy.ParseCourseURL(options.CourseURL)
	if err != nil {
		return download.Summary{}, err
	}
	client, err := courseClient(ctx, options.DataDir, options.ChromeDataDir, ref, logger)
	if err != nil {
		return download.Summary{}, err
	}
	planOptions := udemy.PlanOptions{Quality: udemy.BestQuality, CaptionLanguage: options.CaptionLanguage, DownloadAssets: true}
	plan, err := planCourse(ctx, client, ref, planOptions, logger)
	if err != nil {
		return download.Summary{}, err
	}
	tempDir, err := os.MkdirTemp("", playlistTempPattern)
	if err != nil {
		return download.Summary{}, fmt.Errorf("create HLS playlist dir: %w", err)
	}
	defer os.RemoveAll(tempDir)
	tools := download.Tools{
		Files:     downloadtools.Aria2c{Binary: options.Tools.Aria2c, Output: options.ToolOutput},
		Streams:   downloadtools.YTDLP{Binary: options.Tools.YTDLP, Output: options.ToolOutput},
		Playlists: client,
	}
	return download.Course(ctx, plan, tools, download.Options{OutputDir: options.OutputDir, TempDir: tempDir, Chapters: chapters}, logger, onEvent)
}
