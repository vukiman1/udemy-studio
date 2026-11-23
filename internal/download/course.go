// Package download writes a planned course to disk the way the Python downloader does, so both produce the same tree.
package download

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const (
	dirMode            = 0o750
	aria2ControlSuffix = ".aria2"
)

var errUnsafeName = errors.New("name is not a single path element")

type FileFetcher interface {
	FetchFile(ctx context.Context, url, dir, name string) error
}

type HLSFetcher interface {
	FetchHLS(ctx context.Context, playlistURL, outputTemplate string) error
}

type Tools struct {
	Files     FileFetcher
	Streams   HLSFetcher
	Playlists udemy.PlaylistFetcher
}

type Options struct {
	OutputDir string
	// TempDir holds the HLS playlists yt-dlp reads; the caller removes it.
	TempDir  string
	Chapters ChapterFilter
}

type Summary struct {
	Downloaded     int
	AlreadyPresent int
	Failed         []string
}

type courseDownload struct {
	tools   Tools
	options Options
	logger  *slog.Logger
	onEvent func(jobs.Event)
	summary Summary
}

// Course downloads every selected lecture of the plan; a lecture that fails is recorded and the next one starts.
func Course(ctx context.Context, plan udemy.Plan, tools Tools, options Options, logger *slog.Logger, onEvent func(jobs.Event)) (Summary, error) {
	courseDir, err := makeDir(options.OutputDir, plan.CourseDir)
	if err != nil {
		return Summary{}, err
	}
	if err := writeManifest(courseDir, plan); err != nil {
		return Summary{}, err
	}
	run := &courseDownload{tools: tools, options: options, logger: logger, onEvent: onEvent}
	onEvent(jobs.Event{Kind: jobs.EventCourseStarted, CourseTitle: plan.CourseName, TotalLectures: options.Chapters.lectureCount(plan)})
	for _, chapter := range plan.Chapters {
		if !options.Chapters.includes(chapter.Index) {
			continue
		}
		chapterDir, err := makeDir(courseDir, chapter.Dir)
		if err != nil {
			return run.summary, err
		}
		for _, lecture := range chapter.Lectures {
			if lecture.Kind == udemy.KindQuiz {
				continue
			}
			if err := run.lecture(ctx, chapterDir, lecture); err != nil {
				return run.summary, err
			}
		}
	}
	onEvent(jobs.Event{Kind: jobs.EventSummary})
	return run.summary, nil
}

func (run *courseDownload) lecture(ctx context.Context, chapterDir string, lecture udemy.LecturePlan) error {
	run.onEvent(jobs.Event{Kind: jobs.EventLectureStarted, LectureTitle: lecture.Title})
	failedBefore := len(run.summary.Failed)
	steps := []func(context.Context, string, udemy.LecturePlan) error{run.mainFile, run.captions, run.assets}
	for _, step := range steps {
		if err := step(ctx, chapterDir, lecture); err != nil {
			return run.abort(ctx, lecture, err)
		}
	}
	if len(run.summary.Failed) > failedBefore {
		run.onEvent(jobs.Event{Kind: jobs.EventLectureFailed, LectureTitle: lecture.Title, Failures: run.summary.Failed[failedBefore:]})
		return nil
	}
	run.onEvent(jobs.Event{Kind: jobs.EventLectureDone, LectureTitle: lecture.Title})
	return nil
}

// abort ends the course like an exception in the Python downloader; a cancel ends it without a report.
func (run *courseDownload) abort(ctx context.Context, lecture udemy.LecturePlan, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	run.onEvent(jobs.Event{Kind: jobs.EventLectureFailed, LectureTitle: lecture.Title, Failures: []string{fmt.Sprintf("%s (%v)", lecture.Title, err)}})
	return fmt.Errorf("download %s: %w", lecture.Title, err)
}

func (run *courseDownload) recordFailure(ctx context.Context, failure string, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	run.logger.Warn("download failed", "item", failure, "err", err)
	run.summary.Failed = append(run.summary.Failed, failure)
	return nil
}

func makeDir(parent, name string) (string, error) {
	if !filepath.IsLocal(name) || filepath.Base(name) != name {
		return "", fmt.Errorf("%w: %q", errUnsafeName, name)
	}
	dir := filepath.Join(parent, name)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	return dir, nil
}

// isDownloaded rejects a file aria2c left partial: it keeps a control file next to it until the download completes.
func isDownloaded(path string) bool {
	_, err := os.Stat(path + aria2ControlSuffix)
	return isFile(path) && errors.Is(err, fs.ErrNotExist)
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
