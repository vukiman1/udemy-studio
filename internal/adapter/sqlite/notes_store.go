package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite/gen"
	"github.com/vukiman1/udemy-studio/internal/notes"
)

type NotesStore struct {
	queries *gen.Queries
}

func NewNotesStore(database *sql.DB) *NotesStore {
	return &NotesStore{queries: gen.New(database)}
}

func (store *NotesStore) Create(ctx context.Context, note notes.Note) (notes.Note, error) {
	row, err := store.queries.CreateNote(ctx, gen.CreateNoteParams{
		LectureKey: note.LectureKey,
		AtSeconds:  note.AtSeconds,
		Body:       note.Body,
		CreatedAt:  formatTime(note.CreatedAt),
		UpdatedAt:  formatTime(note.UpdatedAt),
	})
	if err != nil {
		return notes.Note{}, fmt.Errorf("insert note of %q: %w", note.LectureKey, err)
	}
	return toNote(row)
}

func (store *NotesStore) UpdateBody(ctx context.Context, id int64, body string, updatedAt time.Time) (notes.Note, error) {
	row, err := store.queries.UpdateNoteBody(ctx, gen.UpdateNoteBodyParams{ID: id, Body: body, UpdatedAt: formatTime(updatedAt)})
	if errors.Is(err, sql.ErrNoRows) {
		return notes.Note{}, notes.ErrNotFound
	}
	if err != nil {
		return notes.Note{}, fmt.Errorf("update note %d: %w", id, err)
	}
	return toNote(row)
}

func (store *NotesStore) Delete(ctx context.Context, id int64) error {
	deleted, err := store.queries.DeleteNote(ctx, id)
	if err != nil {
		return fmt.Errorf("delete note %d: %w", id, err)
	}
	if deleted == 0 {
		return notes.ErrNotFound
	}
	return nil
}

func (store *NotesStore) ForLecture(ctx context.Context, lectureKey string) ([]notes.Note, error) {
	rows, err := store.queries.ListLectureNotes(ctx, lectureKey)
	if err != nil {
		return nil, fmt.Errorf("list notes of %q: %w", lectureKey, err)
	}
	return toNotes(rows)
}

func (store *NotesStore) All(ctx context.Context) ([]notes.Note, error) {
	rows, err := store.queries.ListNotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	return toNotes(rows)
}

func (store *NotesStore) RenameKey(ctx context.Context, from, to string) error {
	if err := store.queries.RenameNotesKey(ctx, gen.RenameNotesKeyParams{FromKey: from, ToKey: to}); err != nil {
		return fmt.Errorf("rename notes %q to %q: %w", from, to, err)
	}
	return nil
}

func toNotes(rows []gen.Note) ([]notes.Note, error) {
	converted := make([]notes.Note, 0, len(rows))
	for _, row := range rows {
		note, err := toNote(row)
		if err != nil {
			return nil, err
		}
		converted = append(converted, note)
	}
	return converted, nil
}

func toNote(row gen.Note) (notes.Note, error) {
	createdAt, err := parseTime(row.CreatedAt)
	if err != nil {
		return notes.Note{}, fmt.Errorf("note %d created_at: %w", row.ID, err)
	}
	updatedAt, err := parseTime(row.UpdatedAt)
	if err != nil {
		return notes.Note{}, fmt.Errorf("note %d updated_at: %w", row.ID, err)
	}
	return notes.Note{
		ID:         row.ID,
		LectureKey: row.LectureKey,
		AtSeconds:  row.AtSeconds,
		Body:       row.Body,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}
