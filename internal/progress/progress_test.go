package progress_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/progress"
	"github.com/vukiman1/udemy-studio/internal/progress/progresstest"
)

type clock struct{ current time.Time }

func (fake *clock) now() time.Time { return fake.current }

func newTestService() (*progress.Service, *clock) {
	fake := &clock{current: time.Date(2026, 11, 23, 8, 0, 0, 0, time.UTC)}
	return progress.NewService(progresstest.NewMemoryStore(), fake.now), fake
}

func TestSaveRecordsPositionAndTime(t *testing.T) {
	service, fake := newTestService()
	record, err := service.Save(t.Context(), progress.Report{LectureKey: "c/01/001 A", PositionSeconds: 42.5, DurationSeconds: 120})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if record.PositionSeconds != 42.5 || record.DurationSeconds != 120 || !record.UpdatedAt.Equal(fake.current) || record.IsCompleted() {
		t.Errorf("record = %+v", record)
	}
}

func TestCompletionIsStampedOnceAndSurvivesRewatching(t *testing.T) {
	service, fake := newTestService()
	ctx := t.Context()
	completedAt := fake.current
	if _, err := service.Save(ctx, progress.Report{LectureKey: "k", PositionSeconds: 118, DurationSeconds: 120, Completed: true}); err != nil {
		t.Fatalf("complete: %v", err)
	}
	fake.current = fake.current.Add(time.Hour)
	record, err := service.Save(ctx, progress.Report{LectureKey: "k", PositionSeconds: 10, DurationSeconds: 120})
	if err != nil {
		t.Fatalf("rewatch: %v", err)
	}
	if !record.IsCompleted() || !record.CompletedAt.Equal(completedAt) || record.PositionSeconds != 10 {
		t.Errorf("record = %+v, want completed at %v with new position", record, completedAt)
	}
}

func TestSaveRejectsInvalidReports(t *testing.T) {
	cases := map[string]progress.Report{
		"empty key":             {PositionSeconds: 1, DurationSeconds: 10},
		"zero duration":         {LectureKey: "k", PositionSeconds: 0, DurationSeconds: 0},
		"negative position":     {LectureKey: "k", PositionSeconds: -1, DurationSeconds: 10},
		"position past the end": {LectureKey: "k", PositionSeconds: 12, DurationSeconds: 10},
		"not a number":          {LectureKey: "k", PositionSeconds: math.NaN(), DurationSeconds: 10},
		"infinite duration":     {LectureKey: "k", PositionSeconds: 1, DurationSeconds: math.Inf(1)},
	}
	for name, report := range cases {
		t.Run(name, func(t *testing.T) {
			service, _ := newTestService()
			if _, err := service.Save(t.Context(), report); !errors.Is(err, progress.ErrInvalidReport) {
				t.Errorf("Save(%+v) error = %v, want progress.ErrInvalidReport", report, err)
			}
		})
	}
}

func TestSaveAcceptsPositionWithinToleranceOfTheEnd(t *testing.T) {
	service, _ := newTestService()
	if _, err := service.Save(t.Context(), progress.Report{LectureKey: "k", PositionSeconds: 120.4, DurationSeconds: 120}); err != nil {
		t.Errorf("Save near the end: %v", err)
	}
}

func TestAllIndexesRecordsByLectureKey(t *testing.T) {
	service, _ := newTestService()
	ctx := t.Context()
	for _, key := range []string{"a", "b"} {
		if _, err := service.Save(ctx, progress.Report{LectureKey: key, PositionSeconds: 1, DurationSeconds: 10}); err != nil {
			t.Fatalf("Save %s: %v", key, err)
		}
	}
	byKey, err := service.All(ctx)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(byKey) != 2 || byKey["a"].LectureKey != "a" || byKey["b"].LectureKey != "b" {
		t.Errorf("All = %+v", byKey)
	}
}
