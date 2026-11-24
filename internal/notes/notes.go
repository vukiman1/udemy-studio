// Package notes keeps what a learner writes down at a moment of a lecture.
package notes

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const maxBodyRunes = 10_000

var lineBreaks = strings.NewReplacer("\r\n", "\n", "\r", "\n")

var (
	ErrInvalidNote = errors.New("invalid note")
	ErrNotFound    = errors.New("note not found")
)

type Note struct {
	ID         int64
	LectureKey string
	AtSeconds  float64
	Body       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Store interface {
	Create(ctx context.Context, note Note) (Note, error)
	// UpdateBody returns ErrNotFound when no note has the id.
	UpdateBody(ctx context.Context, id int64, body string, updatedAt time.Time) (Note, error)
	// Delete returns ErrNotFound when no note has the id.
	Delete(ctx context.Context, id int64) error
	ForLecture(ctx context.Context, lectureKey string) ([]Note, error)
	All(ctx context.Context) ([]Note, error)
	RenameKey(ctx context.Context, from, to string) error
}

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store, now func() time.Time) *Service {
	return &Service{store: store, now: now}
}

func (service *Service) Add(ctx context.Context, lectureKey string, atSeconds float64, body string) (Note, error) {
	body, err := validBody(body)
	if err != nil {
		return Note{}, err
	}
	if math.IsNaN(atSeconds) || math.IsInf(atSeconds, 0) || atSeconds < 0 {
		return Note{}, fmt.Errorf("%w: moment %v must be a finite number of seconds from the start", ErrInvalidNote, atSeconds)
	}
	now := service.now().UTC()
	note, err := service.store.Create(ctx, Note{LectureKey: lectureKey, AtSeconds: atSeconds, Body: body, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		return Note{}, fmt.Errorf("add a note to %q: %w", lectureKey, err)
	}
	return note, nil
}

func (service *Service) Edit(ctx context.Context, id int64, body string) (Note, error) {
	body, err := validBody(body)
	if err != nil {
		return Note{}, err
	}
	note, err := service.store.UpdateBody(ctx, id, body, service.now().UTC())
	if err != nil {
		return Note{}, fmt.Errorf("edit note %d: %w", id, err)
	}
	return note, nil
}

func (service *Service) Delete(ctx context.Context, id int64) error {
	if err := service.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete note %d: %w", id, err)
	}
	return nil
}

func (service *Service) ForLecture(ctx context.Context, lectureKey string) ([]Note, error) {
	notes, err := service.store.ForLecture(ctx, lectureKey)
	if err != nil {
		return nil, fmt.Errorf("list notes of %q: %w", lectureKey, err)
	}
	return notes, nil
}

func (service *Service) ByLecture(ctx context.Context) (map[string][]Note, error) {
	notes, err := service.store.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	byLecture := map[string][]Note{}
	for _, note := range notes {
		byLecture[note.LectureKey] = append(byLecture[note.LectureKey], note)
	}
	return byLecture, nil
}

// AdoptKeys moves notes written under old lecture keys to their new keys; running it again changes nothing.
func (service *Service) AdoptKeys(ctx context.Context, renames map[string]string) error {
	for _, from := range slices.Sorted(maps.Keys(renames)) {
		if err := service.store.RenameKey(ctx, from, renames[from]); err != nil {
			return fmt.Errorf("move notes of %q to %q: %w", from, renames[from], err)
		}
	}
	return nil
}

func validBody(body string) (string, error) {
	trimmed := strings.TrimSpace(lineBreaks.Replace(body))
	switch length := utf8.RuneCountInString(trimmed); {
	case length == 0:
		return "", fmt.Errorf("%w: the note is empty", ErrInvalidNote)
	case length > maxBodyRunes:
		return "", fmt.Errorf("%w: the note has %d characters, more than %d", ErrInvalidNote, length, maxBodyRunes)
	}
	return trimmed, nil
}
