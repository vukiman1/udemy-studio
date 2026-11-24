package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func TestNoteRequestsReportBadInputAndMissingTargets(t *testing.T) {
	app := newTestApp(t)
	notesPath := "/api/lectures/" + app.lectureID(t, 1) + "/notes"
	t.Run("blank note", func(t *testing.T) {
		assertProblem(t, app.do(t, http.MethodPost, notesPath, gen.NoteDraft{AtSeconds: 1, Body: " \n "}), http.StatusBadRequest)
	})
	t.Run("negative moment", func(t *testing.T) {
		assertProblem(t, app.do(t, http.MethodPost, notesPath, gen.NoteDraft{AtSeconds: -1, Body: "x"}), http.StatusBadRequest)
	})
	t.Run("unknown lecture", func(t *testing.T) {
		assertProblem(t, app.do(t, http.MethodPost, "/api/lectures/nope/notes", gen.NoteDraft{Body: "x"}), http.StatusNotFound)
	})
	t.Run("edit a missing note", func(t *testing.T) {
		assertProblem(t, app.do(t, http.MethodPatch, "/api/notes/404", gen.NoteEdit{Body: "x"}), http.StatusNotFound)
	})
	t.Run("blank edit", func(t *testing.T) {
		added := decode[gen.Note](t, app.do(t, http.MethodPost, notesPath, gen.NoteDraft{AtSeconds: 1, Body: "x"}))
		assertProblem(t, app.do(t, http.MethodPatch, "/api/notes/"+strconv.FormatInt(added.Id, 10), gen.NoteEdit{Body: " "}), http.StatusBadRequest)
	})
	t.Run("delete a missing note", func(t *testing.T) {
		assertProblem(t, app.do(t, http.MethodDelete, "/api/notes/404", nil), http.StatusNotFound)
	})
}

func TestCourseNotesExportAsAMarkdownDownload(t *testing.T) {
	app := newTestApp(t)
	app.do(t, http.MethodPost, "/api/lectures/"+app.lectureID(t, 1)+"/notes", gen.NoteDraft{AtSeconds: 75, Body: "Bắt tay"})
	result := app.do(t, http.MethodGet, "/api/courses/"+app.library.Courses[0].ID+"/notes.md", nil)
	if result.status != http.StatusOK || !strings.Contains(string(result.body), "- **01:15** Bắt tay") ||
		!strings.HasPrefix(result.header.Get("Content-Disposition"), "attachment;") ||
		result.header.Get("Content-Type") != "text/markdown; charset=utf-8" {
		t.Errorf("status=%d headers=%v body=%q", result.status, result.header, result.body)
	}
}
