package download

import (
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/library"
)

const coverURL = "https://img-c.udemycdn.com/course/750x422/1_a.jpg"

func downloadWithCover(t *testing.T, tools *fakeTools, outputDir, cover string) error {
	t.Helper()
	options := Options{OutputDir: outputDir, TempDir: t.TempDir(), Chapters: ChapterFilter{{first: 2, last: 2}}, CoverURL: cover}
	_, err := Course(t.Context(), samplePlan(), Tools{Files: tools, Streams: tools, Playlists: tools, Quizzes: tools}, options, slog.New(slog.DiscardHandler), func(jobs.Event) {})
	return err
}

func TestTheCourseCoverIsFetchedOnceBesideTheChapters(t *testing.T) {
	outputDir := t.TempDir()
	tools := &fakeTools{}
	for range 2 {
		if err := downloadWithCover(t, tools, outputDir, coverURL); err != nil {
			t.Fatalf("Course: %v", err)
		}
	}
	if _, err := os.Stat(filepath.Join(outputDir, samplePlan().CourseDir, library.CoverFileName)); err != nil {
		t.Errorf("cover not beside the chapters: %v", err)
	}
	if fetches := len(slices.DeleteFunc(slices.Clone(tools.fetched), func(name string) bool { return name != library.CoverFileName })); fetches != 1 {
		t.Errorf("cover fetched %d times over two runs; want once", fetches)
	}
}

func TestACoverThatCannotBeFetchedDoesNotFailTheCourse(t *testing.T) {
	if err := downloadWithCover(t, &fakeTools{}, t.TempDir(), failingURL); err != nil {
		t.Errorf("Course = %v; want the course downloaded without its cover", err)
	}
}
