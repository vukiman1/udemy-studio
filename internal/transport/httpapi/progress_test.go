package httpapi

import (
	"net/http"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func TestPutProgressStoresAndReturnsIt(t *testing.T) {
	app := newTestApp(t)
	result := app.do(t, http.MethodPut, "/api/lectures/"+app.lectureID(t, 1)+"/progress", gen.ProgressReport{PositionSeconds: 2, DurationSeconds: 3})
	saved := decode[gen.LectureProgress](t, result)
	if result.status != http.StatusOK || saved.PositionSeconds != 2 || saved.Completed {
		t.Errorf("status=%d saved=%+v", result.status, saved)
	}
}

func TestPutProgressRejectsBadInput(t *testing.T) {
	app := newTestApp(t)
	path := "/api/lectures/" + app.lectureID(t, 1) + "/progress"
	t.Run("position past the end", func(t *testing.T) {
		assertProblem(t, app.do(t, http.MethodPut, path, gen.ProgressReport{PositionSeconds: 99, DurationSeconds: 3}), http.StatusBadRequest)
	})
	t.Run("body is not json", func(t *testing.T) {
		assertProblem(t, app.do(t, http.MethodPut, path, "not an object"), http.StatusBadRequest)
	})
	t.Run("unknown lecture", func(t *testing.T) {
		assertProblem(t, app.do(t, http.MethodPut, "/api/lectures/nope/progress", gen.ProgressReport{PositionSeconds: 1, DurationSeconds: 3}), http.StatusNotFound)
	})
}
