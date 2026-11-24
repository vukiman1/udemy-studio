package notestest

import (
	"context"
	"time"

	"github.com/vukiman1/udemy-studio/internal/notes"
)

type MemoryStore struct {
	notes  map[int64]notes.Note
	nextID int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{notes: map[int64]notes.Note{}, nextID: 1}
}

func (store *MemoryStore) Create(_ context.Context, note notes.Note) (notes.Note, error) {
	note.ID = store.nextID
	store.nextID++
	store.notes[note.ID] = note
	return note, nil
}

func (store *MemoryStore) UpdateBody(_ context.Context, id int64, body string, updatedAt time.Time) (notes.Note, error) {
	note, ok := store.notes[id]
	if !ok {
		return notes.Note{}, notes.ErrNotFound
	}
	note.Body, note.UpdatedAt = body, updatedAt
	store.notes[id] = note
	return note, nil
}

func (store *MemoryStore) Delete(_ context.Context, id int64) error {
	if _, ok := store.notes[id]; !ok {
		return notes.ErrNotFound
	}
	delete(store.notes, id)
	return nil
}

func (store *MemoryStore) ForLecture(_ context.Context, lectureKey string) ([]notes.Note, error) {
	var found []notes.Note
	for _, note := range store.notes {
		if note.LectureKey == lectureKey {
			found = append(found, note)
		}
	}
	return found, nil
}

func (store *MemoryStore) All(_ context.Context) ([]notes.Note, error) {
	all := make([]notes.Note, 0, len(store.notes))
	for _, note := range store.notes {
		all = append(all, note)
	}
	return all, nil
}

func (store *MemoryStore) RenameKey(_ context.Context, from, to string) error {
	for id, note := range store.notes {
		if note.LectureKey == from {
			note.LectureKey = to
			store.notes[id] = note
		}
	}
	return nil
}
