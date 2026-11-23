package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/vukiman1/udemy-studio/internal/jobs"
)

func TestJobStoreRoundTripsAndUpdatesJobs(t *testing.T) {
	database, err := Open(t.Context(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	store := NewJobStore(database)

	created := time.Date(2026, 11, 23, 8, 0, 0, 0, time.UTC)
	started := created.Add(time.Second)
	older := jobs.Job{ID: "a", Request: jobs.Request{CourseURL: "https://acme.udemy.com/course/x/"}, Status: jobs.StatusQueued, CreatedAt: created}
	newer := jobs.Job{
		ID: "b", Request: jobs.Request{CourseURL: "https://acme.udemy.com/course/y/", Chapters: "1-2"}, Status: jobs.StatusRunning,
		CreatedAt: created.Add(time.Minute), StartedAt: &started,
		Progress: jobs.Progress{CourseTitle: "Y", TotalLectures: 4, FinishedCount: 1, CurrentLecture: "002 B"},
		Failures: []string{"001 A (caption: en)"},
	}
	for _, job := range []jobs.Job{older, newer} {
		if err := store.Save(t.Context(), job); err != nil {
			t.Fatalf("Save %s: %v", job.ID, err)
		}
	}
	older.Status, older.Error = jobs.StatusFailed, "boom"
	if err := store.Save(t.Context(), older); err != nil {
		t.Fatalf("update: %v", err)
	}

	listed, err := store.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if diff := cmp.Diff([]jobs.Job{newer, older}, listed); diff != "" {
		t.Errorf("jobs mismatch, newest first (-want +got):\n%s", diff)
	}
}
