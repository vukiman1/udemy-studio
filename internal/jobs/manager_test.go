package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	validCourseURL = "https://acme.udemy.com/course/backend-fundamentals/"
	waitTimeout    = 2 * time.Second
)

type memoryStore struct {
	mutex sync.Mutex
	jobs  map[string]Job
}

func (store *memoryStore) Save(_ context.Context, job Job) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.jobs[job.ID] = job
	return nil
}

func (store *memoryStore) get(id string) Job {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	return store.jobs[id]
}

func (store *memoryStore) List(context.Context) ([]Job, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	jobs := make([]Job, 0, len(store.jobs))
	for _, job := range store.jobs {
		jobs = append(jobs, job)
	}
	return jobs, nil
}

type tokens struct{ err error }

func (source tokens) Token(context.Context, string) (secret.Secret, error) {
	return secret.New("token"), source.err
}

type scriptedDownloader struct {
	events     []Event
	err        error
	blockUntil bool
	gotToken   chan string
	started    chan struct{}
}

func (downloader *scriptedDownloader) Download(ctx context.Context, _ Job, token secret.Secret, onEvent func(Event)) error {
	if downloader.gotToken != nil {
		downloader.gotToken <- token.Reveal()
	}
	for _, event := range downloader.events {
		onEvent(event)
	}
	if downloader.started != nil {
		close(downloader.started)
	}
	if downloader.blockUntil {
		<-ctx.Done()
		return ctx.Err()
	}
	return downloader.err
}

type harness struct {
	manager   *Manager
	store     *memoryStore
	succeeded chan Job
	cancel    context.CancelFunc
	done      chan struct{}
}

func start(t *testing.T, downloader Downloader, tokenErr error) harness {
	t.Helper()
	store := &memoryStore{jobs: map[string]Job{}}
	succeeded := make(chan Job, 1)
	manager := NewManager(Deps{
		Downloader:  downloader,
		Tokens:      tokens{err: tokenErr},
		Store:       store,
		Logger:      slog.New(slog.DiscardHandler),
		Now:         time.Now,
		OnSucceeded: func(_ context.Context, job Job) { succeeded <- job },
	})
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		_ = manager.Run(ctx)
		close(done)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})
	return harness{manager: manager, store: store, succeeded: succeeded, cancel: cancel, done: done}
}

func waitForStatus(t *testing.T, manager *Manager, id string, want Status) Job {
	t.Helper()
	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		for _, job := range manager.List() {
			if job.ID == id && job.Status == want {
				return job
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job %s never reached %s; jobs: %+v", id, want, manager.List())
	return Job{}
}

func TestJobRunsToSuccessAndTracksProgress(t *testing.T) {
	downloader := &scriptedDownloader{
		gotToken: make(chan string, 1),
		events: []Event{
			{Kind: EventCourseStarted, CourseTitle: "Backend", TotalLectures: 2},
			{Kind: EventLectureStarted, LectureTitle: "001 A"},
			{Kind: EventLectureDone, LectureTitle: "001 A"},
			{Kind: EventLectureStarted, LectureTitle: "002 B"},
			{Kind: EventLectureFailed, LectureTitle: "002 B", Failures: []string{"002 B (caption: en)"}},
			{Kind: EventSummary},
		},
	}
	h := start(t, downloader, nil)
	job, err := h.manager.Submit(t.Context(), Request{CourseURL: validCourseURL, Chapters: "1,3-4"})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	finished := waitForStatus(t, h.manager, job.ID, StatusSucceeded)
	if <-downloader.gotToken != "token" {
		t.Error("downloader did not receive the token")
	}
	want := Progress{CourseTitle: "Backend", TotalLectures: 2, FinishedCount: 2}
	if finished.Progress != want || len(finished.Failures) != 1 || finished.StartedAt == nil || finished.FinishedAt == nil {
		t.Errorf("finished = %+v", finished)
	}
	if stored := h.store.get(job.ID); stored.Status != StatusSucceeded {
		t.Errorf("stored status = %s", stored.Status)
	}
	if (<-h.succeeded).ID != job.ID {
		t.Error("success hook did not run for the job")
	}
}

func TestJobFailsWithTheDownloaderOrLoginError(t *testing.T) {
	cases := map[string]struct {
		downloader *scriptedDownloader
		tokenErr   error
		wantError  string
	}{
		"downloader exits with an error": {downloader: &scriptedDownloader{err: errors.New("python exited with status 1")}, wantError: "python exited with status 1"},
		"login is missing":               {downloader: &scriptedDownloader{}, tokenErr: errors.New("not logged in to Udemy: log in to acme.udemy.com"), wantError: "not logged in to Udemy: log in to acme.udemy.com"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			h := start(t, testCase.downloader, testCase.tokenErr)
			job, err := h.manager.Submit(t.Context(), Request{CourseURL: validCourseURL})
			if err != nil {
				t.Fatalf("Submit: %v", err)
			}
			if failed := waitForStatus(t, h.manager, job.ID, StatusFailed); failed.Error != testCase.wantError {
				t.Errorf("error = %q, want %q", failed.Error, testCase.wantError)
			}
		})
	}
}

func TestCancellingARunningJobStopsTheDownload(t *testing.T) {
	downloader := &scriptedDownloader{blockUntil: true, started: make(chan struct{})}
	h := start(t, downloader, nil)
	job, _ := h.manager.Submit(t.Context(), Request{CourseURL: validCourseURL})
	<-downloader.started
	if err := h.manager.Cancel(t.Context(), job.ID); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	waitForStatus(t, h.manager, job.ID, StatusCancelled)
	if err := h.manager.Cancel(t.Context(), job.ID); !errors.Is(err, ErrJobFinished) {
		t.Errorf("second cancel = %v, want ErrJobFinished", err)
	}
}

func TestCancellingAQueuedJobNeverStartsIt(t *testing.T) {
	blocker := &scriptedDownloader{blockUntil: true, started: make(chan struct{})}
	h := start(t, blocker, nil)
	first, _ := h.manager.Submit(t.Context(), Request{CourseURL: validCourseURL})
	<-blocker.started
	second, _ := h.manager.Submit(t.Context(), Request{CourseURL: validCourseURL})
	if err := h.manager.Cancel(t.Context(), second.ID); err != nil {
		t.Fatalf("Cancel queued: %v", err)
	}
	waitForStatus(t, h.manager, second.ID, StatusCancelled)
	_ = h.manager.Cancel(t.Context(), first.ID)
	cancelled := waitForStatus(t, h.manager, second.ID, StatusCancelled)
	if cancelled.StartedAt != nil {
		t.Error("a cancelled queued job was started")
	}
}

func TestSubmitRejectsRequestsTheDownloaderCannotRun(t *testing.T) {
	h := start(t, &scriptedDownloader{}, nil)
	for name, request := range map[string]Request{
		"not a Udemy host":     {CourseURL: "https://example.com/course/x/"},
		"not a course page":    {CourseURL: "https://acme.udemy.com/home/my-courses/"},
		"plain http":           {CourseURL: "http://acme.udemy.com/course/x/"},
		"chapter list garbage": {CourseURL: validCourseURL, Chapters: "1; rm -rf /"},
	} {
		if _, err := h.manager.Submit(t.Context(), request); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("%s: error = %v, want ErrInvalidRequest", name, err)
		}
	}
	if err := h.manager.Cancel(t.Context(), "missing"); !errors.Is(err, ErrJobNotFound) {
		t.Errorf("cancel missing = %v, want ErrJobNotFound", err)
	}
}

func TestRestoreMarksInterruptedJobsAndRequeuesWaitingOnes(t *testing.T) {
	store := &memoryStore{jobs: map[string]Job{
		"interrupted": {ID: "interrupted", Request: Request{CourseURL: validCourseURL}, Status: StatusRunning, CreatedAt: time.Now().Add(-time.Hour)},
		"waiting":     {ID: "waiting", Request: Request{CourseURL: validCourseURL}, Status: StatusQueued, CreatedAt: time.Now()},
	}}
	manager := NewManager(Deps{Downloader: &scriptedDownloader{}, Tokens: tokens{}, Store: store, Logger: slog.New(slog.DiscardHandler), Now: time.Now})
	if err := manager.Restore(t.Context()); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go func() { _ = manager.Run(ctx) }()
	if job := waitForStatus(t, manager, "interrupted", StatusFailed); job.Error != interruptedMessage {
		t.Errorf("interrupted job error = %q", job.Error)
	}
	waitForStatus(t, manager, "waiting", StatusSucceeded)
}

func TestSubscribersSeeEveryStateChange(t *testing.T) {
	h := start(t, &scriptedDownloader{}, nil)
	updates, unsubscribe := h.manager.Subscribe()
	defer unsubscribe()
	job, _ := h.manager.Submit(t.Context(), Request{CourseURL: validCourseURL})
	var seen []Status
	for len(seen) < 3 {
		select {
		case update := <-updates:
			if update.ID == job.ID {
				seen = append(seen, update.Status)
			}
		case <-time.After(waitTimeout):
			t.Fatalf("saw only %v", seen)
		}
	}
	if seen[0] != StatusQueued || seen[1] != StatusRunning || seen[2] != StatusSucceeded {
		t.Errorf("statuses = %v", seen)
	}
}

func TestSubmittingManyJobsNeverBlocks(t *testing.T) {
	blocker := &scriptedDownloader{blockUntil: true, started: make(chan struct{})}
	h := start(t, blocker, nil)
	submitted := make(chan struct{})
	go func() {
		defer close(submitted)
		for range 200 {
			if _, err := h.manager.Submit(t.Context(), Request{CourseURL: validCourseURL}); err != nil {
				t.Errorf("Submit: %v", err)
				return
			}
		}
	}()
	select {
	case <-submitted:
	case <-time.After(waitTimeout):
		t.Fatal("Submit blocked while many jobs were waiting")
	}
}

func TestRestoreWithManyWaitingJobsDoesNotBlock(t *testing.T) {
	store := &memoryStore{jobs: map[string]Job{}}
	for index := range 200 {
		id := fmt.Sprintf("job-%03d", index)
		store.jobs[id] = Job{ID: id, Request: Request{CourseURL: validCourseURL}, Status: StatusQueued, CreatedAt: time.Now().Add(time.Duration(index) * time.Millisecond)}
	}
	manager := NewManager(Deps{Downloader: &scriptedDownloader{}, Tokens: tokens{}, Store: store, Logger: slog.New(slog.DiscardHandler), Now: time.Now})
	restored := make(chan error, 1)
	go func() { restored <- manager.Restore(t.Context()) }()
	select {
	case err := <-restored:
		if err != nil {
			t.Fatalf("Restore: %v", err)
		}
	case <-time.After(waitTimeout):
		t.Fatal("Restore blocked on many waiting jobs")
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go func() { _ = manager.Run(ctx) }()
	waitForStatus(t, manager, "job-000", StatusSucceeded)
	waitForStatus(t, manager, "job-199", StatusSucceeded)
}

func TestFinishedJobsNoLongerShowACurrentLecture(t *testing.T) {
	downloader := &scriptedDownloader{blockUntil: true, started: make(chan struct{}), events: []Event{{Kind: EventLectureStarted, LectureTitle: "005 Intro"}}}
	h := start(t, downloader, nil)
	job, _ := h.manager.Submit(t.Context(), Request{CourseURL: validCourseURL})
	<-downloader.started
	_ = h.manager.Cancel(t.Context(), job.ID)
	if cancelled := waitForStatus(t, h.manager, job.ID, StatusCancelled); cancelled.Progress.CurrentLecture != "" {
		t.Errorf("cancelled job still shows %q", cancelled.Progress.CurrentLecture)
	}

	store := &memoryStore{jobs: map[string]Job{
		"interrupted": {ID: "interrupted", Status: StatusRunning, Progress: Progress{CurrentLecture: "005 Intro"}},
	}}
	manager := NewManager(Deps{Downloader: &scriptedDownloader{}, Tokens: tokens{}, Store: store, Logger: slog.New(slog.DiscardHandler), Now: time.Now})
	if err := manager.Restore(t.Context()); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if restored := manager.List()[0]; restored.Progress.CurrentLecture != "" {
		t.Errorf("interrupted job still shows %q", restored.Progress.CurrentLecture)
	}
}

type lateCancelDownloader struct {
	started chan struct{}
	release chan struct{}
}

func (downloader lateCancelDownloader) Download(context.Context, Job, secret.Secret, func(Event)) error {
	close(downloader.started)
	<-downloader.release
	return nil
}

func TestADownloadThatFinishedIsNotRecordedAsCancelled(t *testing.T) {
	downloader := lateCancelDownloader{started: make(chan struct{}), release: make(chan struct{})}
	h := start(t, downloader, nil)
	job, _ := h.manager.Submit(t.Context(), Request{CourseURL: validCourseURL})
	<-downloader.started
	if err := h.manager.Cancel(t.Context(), job.ID); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	close(downloader.release)
	waitForStatus(t, h.manager, job.ID, StatusSucceeded)
}

func TestShutdownLeavesWaitingJobsQueued(t *testing.T) {
	blocker := &scriptedDownloader{blockUntil: true, started: make(chan struct{})}
	h := start(t, blocker, nil)
	running, _ := h.manager.Submit(t.Context(), Request{CourseURL: validCourseURL})
	<-blocker.started
	waiting, _ := h.manager.Submit(t.Context(), Request{CourseURL: validCourseURL})
	h.cancel()
	<-h.done
	waitForStatus(t, h.manager, running.ID, StatusFailed)
	if job := waitForStatus(t, h.manager, waiting.ID, StatusQueued); job.StartedAt != nil {
		t.Error("a waiting job was started after shutdown")
	}
}

func idleManager(store Store) *Manager {
	return NewManager(Deps{Downloader: &scriptedDownloader{}, Tokens: tokens{}, Store: store, Logger: slog.New(slog.DiscardHandler), Now: time.Now})
}

func TestSubscriberThatFallsBehindIsClosedSoItReloads(t *testing.T) {
	manager := idleManager(&memoryStore{jobs: map[string]Job{}})
	updates, unsubscribe := manager.Subscribe()
	defer unsubscribe()
	for range subscriberBuffer + 1 {
		if _, err := manager.Submit(t.Context(), Request{CourseURL: validCourseURL}); err != nil {
			t.Fatalf("Submit: %v", err)
		}
	}
	for range subscriberBuffer {
		<-updates
	}
	select {
	case _, isOpen := <-updates:
		if isOpen {
			t.Error("subscriber got an update past its buffer instead of being closed")
		}
	case <-time.After(waitTimeout):
		t.Fatal("subscriber that fell behind was neither closed nor updated")
	}
}
