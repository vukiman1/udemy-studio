package progress

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"time"
)

const positionToleranceSeconds = 1

var ErrInvalidReport = errors.New("invalid progress report")

type Record struct {
	LectureKey      string
	PositionSeconds float64
	DurationSeconds float64
	CompletedAt     *time.Time
	UpdatedAt       time.Time
}

func (record Record) IsCompleted() bool { return record.CompletedAt != nil }

type Report struct {
	LectureKey      string
	PositionSeconds float64
	DurationSeconds float64
	Completed       bool
}

type Store interface {
	Find(ctx context.Context, lectureKey string) (Record, bool, error)
	Upsert(ctx context.Context, record Record) error
	All(ctx context.Context) ([]Record, error)
	// RenameKey moves from's record to to, and keeps it under from if to already has one.
	RenameKey(ctx context.Context, from, to string) error
}

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store, now func() time.Time) *Service {
	return &Service{store: store, now: now}
}

func (service *Service) Save(ctx context.Context, report Report) (Record, error) {
	if err := validate(report); err != nil {
		return Record{}, err
	}
	previous, found, err := service.store.Find(ctx, report.LectureKey)
	if err != nil {
		return Record{}, fmt.Errorf("load progress of %q: %w", report.LectureKey, err)
	}
	now := service.now().UTC()
	record := Record{
		LectureKey:      report.LectureKey,
		PositionSeconds: report.PositionSeconds,
		DurationSeconds: report.DurationSeconds,
		UpdatedAt:       now,
	}
	switch {
	case found && previous.IsCompleted():
		record.CompletedAt = previous.CompletedAt
	case report.Completed:
		record.CompletedAt = &now
	}
	if err := service.store.Upsert(ctx, record); err != nil {
		return Record{}, fmt.Errorf("save progress of %q: %w", report.LectureKey, err)
	}
	return record, nil
}

func (service *Service) All(ctx context.Context) (map[string]Record, error) {
	records, err := service.store.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list progress: %w", err)
	}
	byKey := make(map[string]Record, len(records))
	for _, record := range records {
		byKey[record.LectureKey] = record
	}
	return byKey, nil
}

// AdoptKeys moves progress saved under old lecture keys to their new keys; running it again changes nothing.
func (service *Service) AdoptKeys(ctx context.Context, renames map[string]string) error {
	for _, from := range slices.Sorted(maps.Keys(renames)) {
		if err := service.store.RenameKey(ctx, from, renames[from]); err != nil {
			return fmt.Errorf("move progress of %q to %q: %w", from, renames[from], err)
		}
	}
	return nil
}

func validate(report Report) error {
	switch {
	case report.LectureKey == "":
		return fmt.Errorf("%w: empty lecture key", ErrInvalidReport)
	case !isFinite(report.PositionSeconds) || !isFinite(report.DurationSeconds):
		return fmt.Errorf("%w: position and duration must be finite", ErrInvalidReport)
	case report.DurationSeconds <= 0:
		return fmt.Errorf("%w: duration %.3f must be positive", ErrInvalidReport, report.DurationSeconds)
	case report.PositionSeconds < 0 || report.PositionSeconds > report.DurationSeconds+positionToleranceSeconds:
		return fmt.Errorf("%w: position %.3f outside 0..%.3f", ErrInvalidReport, report.PositionSeconds, report.DurationSeconds)
	}
	return nil
}

func isFinite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
