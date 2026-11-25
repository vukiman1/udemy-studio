package httpapi

import (
	"context"
	"fmt"

	"github.com/vukiman1/udemy-studio/internal/learning"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

const historyLimit = 50

func (server apiServer) ListHistory(ctx context.Context, _ gen.ListHistoryRequestObject) (gen.ListHistoryResponseObject, error) {
	records, err := server.deps.Progress.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list history: %w", err)
	}
	entries := learning.History(server.deps.Catalog.Current(), records, historyLimit)
	response := make(gen.ListHistory200JSONResponse, 0, len(entries))
	for _, entry := range entries {
		response = append(response, gen.HistoryEntry{
			CourseId:        entry.Course.ID,
			CourseTitle:     entry.Course.Title,
			LectureId:       entry.Lecture.ID,
			LectureTitle:    entry.Lecture.Title,
			PositionSeconds: entry.Record.PositionSeconds,
			Completed:       entry.Record.IsCompleted(),
			UpdatedAt:       entry.Record.UpdatedAt,
		})
	}
	return response, nil
}
