package progresstest

import (
	"context"

	"github.com/vukiman1/udemy-studio/internal/progress"
)

type MemoryStore struct {
	records map[string]progress.Record
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: map[string]progress.Record{}}
}

func (store *MemoryStore) Find(_ context.Context, lectureKey string) (progress.Record, bool, error) {
	record, ok := store.records[lectureKey]
	return record, ok, nil
}

func (store *MemoryStore) Upsert(_ context.Context, record progress.Record) error {
	store.records[record.LectureKey] = record
	return nil
}

func (store *MemoryStore) All(_ context.Context) ([]progress.Record, error) {
	records := make([]progress.Record, 0, len(store.records))
	for _, record := range store.records {
		records = append(records, record)
	}
	return records, nil
}
