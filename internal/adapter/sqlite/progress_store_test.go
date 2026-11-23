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

func TestRenameKeyMovesProgressButNeverOverwrites(t *testing.T) {
	store := openTestStore(t)
	ctx := t.Context()
	saved := time.Unix(100, 0).UTC()
	for _, record := range []progress.Record{
		{LectureKey: "course/01 - A/001 Moved", PositionSeconds: 5, DurationSeconds: 10, UpdatedAt: saved},
		{LectureKey: "course/01 - A/002 Kept", PositionSeconds: 6, DurationSeconds: 10, UpdatedAt: saved},
		{LectureKey: "udemy-lecture:2", PositionSeconds: 9, DurationSeconds: 10, UpdatedAt: saved},
	} {
		if err := store.Upsert(ctx, record); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
	}
	for from, to := range map[string]string{"course/01 - A/001 Moved": "udemy-lecture:1", "course/01 - A/002 Kept": "udemy-lecture:2"} {
		if err := store.RenameKey(ctx, from, to); err != nil {
			t.Fatalf("RenameKey: %v", err)
		}
	}
	positions := map[string]float64{}
	records, err := store.All(ctx)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	for _, record := range records {
		positions[record.LectureKey] = record.PositionSeconds
	}
	want := map[string]float64{"udemy-lecture:1": 5, "course/01 - A/002 Kept": 6, "udemy-lecture:2": 9}
	if diff := cmp.Diff(want, positions); diff != "" {
		t.Errorf("progress after renames (-want +got):\n%s", diff)
	}
}
