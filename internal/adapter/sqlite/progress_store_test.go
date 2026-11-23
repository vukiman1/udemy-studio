package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/vukiman1/udemy-studio/internal/progress"
)

func openTestStore(t *testing.T) *ProgressStore {
	t.Helper()
	database, err := Open(t.Context(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return NewProgressStore(database)
}

func TestProgressStoreRoundTripsRecords(t *testing.T) {
	store := openTestStore(t)
	ctx := t.Context()
	completedAt := time.Date(2026, 11, 23, 8, 0, 0, 123456789, time.UTC)
	inProgress := progress.Record{LectureKey: "c/01/001 A", PositionSeconds: 12.5, DurationSeconds: 100, UpdatedAt: completedAt.Add(time.Minute)}
	finished := progress.Record{LectureKey: "c/01/002 B", PositionSeconds: 99, DurationSeconds: 100, CompletedAt: &completedAt, UpdatedAt: completedAt}

	for _, record := range []progress.Record{inProgress, finished} {
		if err := store.Upsert(ctx, record); err != nil {
			t.Fatalf("Upsert %s: %v", record.LectureKey, err)
		}
	}
	all, err := store.All(ctx)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if diff := cmp.Diff([]progress.Record{inProgress, finished}, all); diff != "" {
		t.Errorf("All mismatch, newest first (-want +got):\n%s", diff)
	}
}

func TestProgressStoreUpsertReplacesExistingRow(t *testing.T) {
	store := openTestStore(t)
	ctx := t.Context()
	first := progress.Record{LectureKey: "k", PositionSeconds: 1, DurationSeconds: 10, UpdatedAt: time.Unix(100, 0).UTC()}
	second := progress.Record{LectureKey: "k", PositionSeconds: 7, DurationSeconds: 10, UpdatedAt: time.Unix(200, 0).UTC()}
	for _, record := range []progress.Record{first, second} {
		if err := store.Upsert(ctx, record); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
	}
	found, ok, err := store.Find(ctx, "k")
	if err != nil || !ok {
		t.Fatalf("Find = %t, %v", ok, err)
	}
	if diff := cmp.Diff(second, found); diff != "" {
		t.Errorf("Find mismatch (-want +got):\n%s", diff)
	}
}

func TestProgressStoreFindMissingKey(t *testing.T) {
	_, ok, err := openTestStore(t).Find(t.Context(), "missing")
	if err != nil || ok {
		t.Errorf("Find(missing) = %t, %v; want not found without error", ok, err)
	}
}

func TestOpenIsIdempotentOnAnExistingDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")
	for attempt := 1; attempt <= 2; attempt++ {
		database, err := Open(t.Context(), path)
		if err != nil {
			t.Fatalf("Open attempt %d: %v", attempt, err)
		}
		_ = database.Close()
	}
}

func TestProgressStoreNeverClearsACompletion(t *testing.T) {
	store := openTestStore(t)
	ctx := t.Context()
	completedAt := time.Unix(100, 0).UTC()
	completed := progress.Record{LectureKey: "k", PositionSeconds: 10, DurationSeconds: 10, CompletedAt: &completedAt, UpdatedAt: completedAt}
	staleSave := progress.Record{LectureKey: "k", PositionSeconds: 4, DurationSeconds: 10, UpdatedAt: time.Unix(200, 0).UTC()}
	for _, record := range []progress.Record{completed, staleSave} {
		if err := store.Upsert(ctx, record); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
	}
	found, _, err := store.Find(ctx, "k")
	if err != nil || found.CompletedAt == nil || !found.CompletedAt.Equal(completedAt) {
		t.Errorf("CompletedAt = %v, %v; want %v", found.CompletedAt, err, completedAt)
	}
}
