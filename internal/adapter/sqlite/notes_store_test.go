package sqlite

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/notes"
)

func openTestNotes(t *testing.T) *notes.Service {
	t.Helper()
	database, err := Open(t.Context(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return notes.NewService(NewNotesStore(database), func() time.Time { return time.Date(2026, 11, 24, 8, 0, 0, 0, time.UTC) })
}

func TestNotesFollowTheirLectureToItsUdemyID(t *testing.T) {
	service := openTestNotes(t)
	for _, at := range []float64{90, 12.5} {
		if _, err := service.Add(t.Context(), "path:Course/01 - Intro/001 Welcome", at, "note"); err != nil {
			t.Fatalf("Add: %v", err)
		}
	}
	renames := map[string]string{"path:Course/01 - Intro/001 Welcome": "udemy-lecture:42"}
	for range 2 {
		if err := service.AdoptKeys(t.Context(), renames); err != nil {
			t.Fatalf("AdoptKeys: %v", err)
		}
	}
	moved, err := service.ForLecture(t.Context(), "udemy-lecture:42")
	if err != nil || len(moved) != 2 || moved[0].AtSeconds != 12.5 || moved[1].AtSeconds != 90 {
		t.Errorf("notes under the new key = %+v, %v; want both, earliest first", moved, err)
	}
}

func TestEditingOrDeletingAMissingNoteReportsNotFound(t *testing.T) {
	service := openTestNotes(t)
	if _, err := service.Edit(t.Context(), 404, "text"); !errors.Is(err, notes.ErrNotFound) {
		t.Errorf("Edit missing note: %v, want ErrNotFound", err)
	}
	if err := service.Delete(t.Context(), 404); !errors.Is(err, notes.ErrNotFound) {
		t.Errorf("Delete missing note: %v, want ErrNotFound", err)
	}
}
