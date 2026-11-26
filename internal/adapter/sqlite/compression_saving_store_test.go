package sqlite

import (
	"path/filepath"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/transcode"
)

func TestCompressionSavingsAreSummedPerCourseAndOverall(t *testing.T) {
	database, err := Open(t.Context(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	store := NewCompressionSavingStore(database)
	if totals, err := store.AllTotals(t.Context()); err != nil || totals != (transcode.SizeTotals{}) {
		t.Fatalf("empty totals = %+v, %v", totals, err)
	}
	for _, saving := range []transcode.Saving{
		{Path: "a/1.mp4", CourseDir: "a", BytesBefore: 100, BytesAfter: 40},
		{Path: "a/2.mp4", CourseDir: "a", BytesBefore: 100, BytesAfter: 20},
		{Path: "b/1.mp4", CourseDir: "b", BytesBefore: 200, BytesAfter: 100},
		{Path: "a/1.mp4", CourseDir: "a", BytesBefore: 100, BytesAfter: 30},
	} {
		if err := store.Save(t.Context(), saving); err != nil {
			t.Fatalf("Save %s: %v", saving.Path, err)
		}
	}
	course, err := store.CourseTotals(t.Context(), "a")
	if err != nil || course != (transcode.SizeTotals{BytesBefore: 200, BytesAfter: 50}) {
		t.Errorf("course totals = %+v, %v", course, err)
	}
	all, err := store.AllTotals(t.Context())
	if err != nil || all != (transcode.SizeTotals{BytesBefore: 400, BytesAfter: 150}) {
		t.Errorf("all totals = %+v, %v", all, err)
	}
}
