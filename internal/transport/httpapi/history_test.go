package httpapi

import (
	"net/http"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func TestHistoryShowsAWatchedLectureWhereItWasLeft(t *testing.T) {
	app := newTestApp(t)
	lectureID := app.lectureID(t, 1)
	app.do(t, http.MethodPut, "/api/lectures/"+lectureID+"/progress", gen.ProgressReport{PositionSeconds: 2, DurationSeconds: 3})

	result := app.do(t, http.MethodGet, "/api/history", nil)
	history := decode[[]gen.HistoryEntry](t, result)
	if result.status != http.StatusOK || len(history) != 1 || history[0].LectureId != lectureID || history[0].PositionSeconds != 2 || history[0].Completed {
		t.Errorf("status=%d history=%+v, want the lecture left at 2s", result.status, history)
	}
}
