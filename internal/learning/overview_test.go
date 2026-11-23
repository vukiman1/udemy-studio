package learning

import (
	"testing"
	"testing/fstest"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/progress"
)

var baseTime = time.Date(2026, 11, 23, 8, 0, 0, 0, time.UTC)

type fixture struct {
	scanned *library.Library
	byTitle map[string]library.Lecture
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	scanned, err := library.Scan(fstest.MapFS{
		"alpha/01 - One/001 A1.mp4":          {},
		"alpha/01 - One/002 A2.mp4":          {},
		"alpha/02 - Two/001 A3.html":         {},
		"alpha/02 - Two/002 Drm only_en.srt": {},
		"beta/01 - One/001 B1.mp4":           {},
		"beta/01 - One/002 B2.mp4":           {},
	})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	byTitle := map[string]library.Lecture{}
	for _, course := range scanned.Courses {
		for _, chapter := range course.Chapters {
			for _, lecture := range chapter.Lectures {
				byTitle[lecture.Title] = lecture
			}
		}
	}
	return fixture{scanned: scanned, byTitle: byTitle}
}

func (data fixture) record(title string, minutesAfterBase int, completed bool) progress.Record {
	updatedAt := baseTime.Add(time.Duration(minutesAfterBase) * time.Minute)
	record := progress.Record{LectureKey: data.byTitle[title].Key, PositionSeconds: 30, DurationSeconds: 100, UpdatedAt: updatedAt}
	if completed {
		record.CompletedAt = &updatedAt
	}
	return record
}

func index(records ...progress.Record) map[string]progress.Record {
	byKey := map[string]progress.Record{}
	for _, record := range records {
		byKey[record.LectureKey] = record
	}
	return byKey
}

func TestResumePicksTheMostRecentUnfinishedLecture(t *testing.T) {
	data := newFixture(t)
	target, ok := Resume(data.scanned, index(data.record("A1", 1, false), data.record("B1", 5, false)))
	if !ok || target.LectureID != data.byTitle["B1"].ID || target.PositionSeconds != 30 {
		t.Errorf("Resume = %+v, %t; want B1 at 30s", target, ok)
	}
}

func TestResumeMovesOnToTheNextLectureAcrossChapters(t *testing.T) {
	data := newFixture(t)
	target, ok := Resume(data.scanned, index(data.record("A2", 1, true)))
	if !ok || target.LectureID != data.byTitle["A3"].ID || target.PositionSeconds != 0 {
		t.Errorf("Resume = %+v, %t; want A3 from the start", target, ok)
	}
}

func TestResumeHasNothingWhenTheLastPlayableLectureIsDone(t *testing.T) {
	data := newFixture(t)
	if target, ok := Resume(data.scanned, index(data.record("A3", 1, true))); ok {
		t.Errorf("Resume = %+v, want none after the course's last playable lecture", target)
	}
}

func TestResumeHasNothingWithoutProgress(t *testing.T) {
	if target, ok := Resume(newFixture(t).scanned, nil); ok {
		t.Errorf("Resume = %+v, want none", target)
	}
}

func TestResumeIgnoresProgressOfLecturesNoLongerInTheLibrary(t *testing.T) {
	data := newFixture(t)
	stale := progress.Record{LectureKey: "alpha/01 - One/009 Removed", UpdatedAt: baseTime.Add(time.Hour)}
	target, ok := Resume(data.scanned, index(stale, data.record("A1", 1, false)))
	if !ok || target.LectureID != data.byTitle["A1"].ID {
		t.Errorf("Resume = %+v, %t; want A1", target, ok)
	}
}

func TestOverviewsCountOnlyPlayableLectures(t *testing.T) {
	data := newFixture(t)
	overviews := Overviews(data.scanned, index(data.record("A1", 1, true), data.record("A2", 2, false)))
	alpha := overviews[0]
	if alpha.Course.Title != "alpha" || alpha.PlayableLectures != 3 || alpha.CompletedLectures != 1 {
		t.Errorf("alpha = %d/%d playable/completed", alpha.PlayableLectures, alpha.CompletedLectures)
	}
	if alpha.Resume == nil || alpha.Resume.LectureID != data.byTitle["A2"].ID {
		t.Errorf("alpha resume = %+v, want A2", alpha.Resume)
	}
	if beta := overviews[1]; beta.Resume != nil || beta.CompletedLectures != 0 {
		t.Errorf("beta = %+v, want untouched", beta)
	}
}
