package pythondownloader

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	fakeModeEnv     = "FAKE_PYTHON_MODE"
	tokenValue      = "token-that-must-not-reach-argv" //nolint:gosec // reason: fixture value checked for leaks
	downloadTimeout = 10 * time.Second
)

type recorder struct {
	mutex  sync.Mutex
	events []jobs.Event
}

func (recorder *recorder) record(event jobs.Event) {
	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()
	recorder.events = append(recorder.events, event)
}

func (recorder *recorder) snapshot() []jobs.Event {
	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()
	return append([]jobs.Event(nil), recorder.events...)
}

func newDownloader(t *testing.T, mode string) (*Downloader, Config) {
	t.Helper()
	t.Setenv(fakeModeEnv, mode)
	fakePython, err := filepath.Abs("testdata/fake-python.sh")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	config := Config{
		Python:          fakePython,
		RepoDir:         t.TempDir(),
		OutputDir:       t.TempDir(),
		LogDir:          t.TempDir(),
		CaptionLanguage: "en",
	}
	return New(config, slog.New(slog.DiscardHandler)), config
}

func testJob() jobs.Job {
	return jobs.Job{ID: "job1", Request: jobs.Request{CourseURL: "https://acme.udemy.com/course/backend/", Chapters: "4"}}
}

func TestDownloadForwardsEventsAndKeepsTheTokenOutOfArguments(t *testing.T) {
	downloader, config := newDownloader(t, "success")
	events := &recorder{}
	if err := downloader.Download(t.Context(), testJob(), secret.New(tokenValue), events.record); err != nil {
		t.Fatalf("Download: %v", err)
	}
	want := []jobs.Event{
		{Kind: jobs.EventCourseStarted, CourseTitle: "Backend", TotalLectures: 2},
		{Kind: jobs.EventLectureStarted, LectureTitle: "001 A"},
		{Kind: jobs.EventLectureDone, LectureTitle: "001 A"},
		{Kind: jobs.EventLectureFailed, LectureTitle: "002 B", Failures: []string{"002 B (caption: en)"}},
		{Kind: jobs.EventSummary},
	}
	if diff := cmp.Diff(want, events.snapshot()); diff != "" {
		t.Errorf("events mismatch (-want +got):\n%s", diff)
	}

	log, err := os.ReadFile(filepath.Join(config.LogDir, "job1.log"))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	text := string(log)
	repoDir, err := filepath.EvalSymlinks(config.RepoDir)
	if err != nil {
		t.Fatalf("resolve repo dir: %v", err)
	}
	expectations := []string{
		"argv: -m udemy_downloader https://acme.udemy.com/course/backend/ --download-captions --download-assets -l en -o " + config.OutputDir + " --events-json /dev/fd/3 --chapter 4",
		"token present: yes",
		"working dir: " + repoDir,
	}
	for _, expected := range expectations {
		if !strings.Contains(text, expected) {
			t.Errorf("log lacks %q:\n%s", expected, text)
		}
	}
	if strings.Contains(text, tokenValue) {
		t.Error("the token reached the process arguments or output")
	}
}

func TestDownloadReportsTheLastLogLineWhenPythonFails(t *testing.T) {
	downloader, _ := newDownloader(t, "failure")
	err := downloader.Download(t.Context(), testJob(), secret.New(tokenValue), (&recorder{}).record)
	if err == nil || !strings.Contains(err.Error(), "login rejected") {
		t.Errorf("error = %v, want it to quote the failure", err)
	}
}

func TestOversizedEventLineDoesNotHangTheDownload(t *testing.T) {
	downloader, _ := newDownloader(t, "oversized-event")
	ctx, cancel := context.WithTimeout(t.Context(), downloadTimeout)
	defer cancel()
	if err := downloader.Download(ctx, testJob(), secret.New(tokenValue), (&recorder{}).record); err != nil {
		t.Errorf("Download = %v, want it to finish", err)
	}
}
