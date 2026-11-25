package learning

import (
	"slices"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/progress"
)

func TestHistoryListsWatchedLecturesNewestFirstUpToTheLimit(t *testing.T) {
	data := newFixture(t)
	gone := progress.Record{LectureKey: "a lecture since deleted", UpdatedAt: baseTime.Add(time.Hour)}
	records := index(data.record("A1", 1, true), data.record("B2", 3, false), data.record("A3", 2, false), gone)

	entries := History(data.scanned, records, 2)

	titles := make([]string, 0, len(entries))
	for _, entry := range entries {
		titles = append(titles, entry.Course.Title+"/"+entry.Lecture.Title)
	}
	if want := []string{"beta/B2", "alpha/A3"}; !slices.Equal(titles, want) {
		t.Errorf("history = %v, want %v", titles, want)
	}
}
