package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/vukiman1/udemy-studio/internal/adapter/downloadtools"
	"github.com/vukiman1/udemy-studio/internal/adapter/udemyapi"
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

// courseDownload is what both the CLI and a download job ask the downloader for.
type courseDownload struct {
	ref             udemy.CourseRef
	chapters        download.ChapterFilter
	outputDir       string
	captionLanguage string
	tools           ToolPaths
	toolOutput      io.Writer
}

// DownloadCourse plans a course with the selected Chrome profile's login and downloads it as the Python downloader would.
func DownloadCourse(ctx context.Context, options DownloadCourseOptions, logger *slog.Logger, onEvent func(jobs.Event)) (download.Summary, error) {
	request, err := newCourseDownload(options.CourseURL, options.Chapters)
	if err != nil {
		return download.Summary{}, err
	}
	request.outputDir, request.captionLanguage = options.OutputDir, options.CaptionLanguage
	request.tools, request.toolOutput = options.Tools, options.ToolOutput
	client, err := courseClient(ctx, options.DataDir, options.ChromeDataDir, request.ref, logger)
	if err != nil {
		return download.Summary{}, err
	}
	return downloadCourse(ctx, client, request, logger, onEvent)
}

func newCourseDownload(courseURL, chapters string) (courseDownload, error) {
	ref, err := udemy.ParseCourseURL(courseURL)
	if err != nil {
		return courseDownload{}, err
	}
	filter, err := download.ParseChapterFilter(chapters)
	if err != nil {
		return courseDownload{}, err
	}
	return courseDownload{ref: ref, chapters: filter}, nil
}

func downloadCourse(ctx context.Context, client *udemyapi.Client, request courseDownload, logger *slog.Logger, onEvent func(jobs.Event)) (download.Summary, error) {
	stopKeepingAwake := keepAwake(ctx, logger)
	defer stopKeepingAwake()
	planOptions := udemy.PlanOptions{Quality: udemy.BestQuality, CaptionLanguage: request.captionLanguage, DownloadAssets: true}
	plan, err := planCourse(ctx, client, request.ref, planOptions, logger)
	if err != nil {
		return download.Summary{}, err
	}
	tempDir, err := os.MkdirTemp("", playlistTempPattern)
	if err != nil {
		return download.Summary{}, fmt.Errorf("create HLS playlist dir: %w", err)
	}
	defer os.RemoveAll(tempDir)
	tools := download.Tools{
		Files:     downloadtools.Aria2c{Binary: request.tools.Aria2c, Output: request.toolOutput},
		Streams:   downloadtools.YTDLP{Binary: request.tools.YTDLP, Output: request.toolOutput},
		Playlists: client,
		Quizzes:   client,
	}
	options := download.Options{OutputDir: request.outputDir, TempDir: tempDir, Chapters: request.chapters}
	return download.Course(ctx, plan, tools, options, logger, onEvent)
}
