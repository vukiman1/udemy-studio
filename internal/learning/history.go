package learning

import (
	"cmp"
	"slices"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/progress"
)

type HistoryEntry struct {
	Course  library.Course
	Lecture library.Lecture
	Record  progress.Record
}

func History(scanned *library.Library, records map[string]progress.Record, limit int) []HistoryEntry {
	var entries []HistoryEntry
	for _, course := range scanned.Courses {
		for _, lecture := range playableLectures(course) {
			if record, ok := records[lecture.Key]; ok {
				entries = append(entries, HistoryEntry{Course: course, Lecture: lecture, Record: record})
			}
		}
	}
	slices.SortFunc(entries, func(left, right HistoryEntry) int {
		return cmp.Compare(right.Record.UpdatedAt.UnixNano(), left.Record.UpdatedAt.UnixNano())
	})
	return entries[:min(limit, len(entries))]
}
