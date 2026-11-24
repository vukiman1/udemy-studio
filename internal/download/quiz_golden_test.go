package download

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/platform/pyjson"
	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const (
	quizFixtureDirEnv  = "QUIZ_FIXTURE_DIR"
	defaultQuizFixture = "../../testdata/quiz"
	quizPageSuffix     = ".html"
	quizErrorSuffix    = ".error"
)

var pythonQuizErrors = map[string]error{
	"KeyError":   errMissingQuizField,
	"IndexError": errEmptyQuiz,
}

type quizFixture struct {
	Lecture struct {
		ID           int64                `json:"id"`
		LectureTitle string               `json:"lecture_title"`
		Data         udemy.CurriculumItem `json:"data"`
	} `json:"lecture"`
	Assessments json.RawMessage `json:"assessments"`
}

func TestQuizPagesMatchPythonDownloader(t *testing.T) {
	fixtureDir := os.Getenv(quizFixtureDirEnv)
	if fixtureDir == "" {
		fixtureDir = defaultQuizFixture
	}
	fixtures := os.DirFS(fixtureDir)
	names, err := fs.Glob(fixtures, "*.json")
	if err != nil || len(names) == 0 {
		t.Fatalf("no quiz fixtures in %s: %v", fixtureDir, err)
	}
	for _, name := range names {
		stem := strings.TrimSuffix(name, path.Ext(name))
		t.Run(stem, func(t *testing.T) {
			lecture, assessments := loadQuizFixture(t, fixtures, name)
			got, renderErr := quizPage(lecture, assessments)
			want, err := fs.ReadFile(fixtures, stem+quizPageSuffix)
			if errors.Is(err, fs.ErrNotExist) {
				assertPythonQuizError(t, fixtures, stem, renderErr)
				return
			}
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if renderErr != nil || got != string(want) {
				t.Errorf("quizPage error = %v; page differs from Python's (got %d bytes, want %d)", renderErr, len(got), len(want))
			}
		})
	}
}

func loadQuizFixture(t *testing.T, fixtures fs.FS, name string) (udemy.LecturePlan, []pyjson.Value) {
	t.Helper()
	raw, err := fs.ReadFile(fixtures, name)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture quizFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	decoded, err := pyjson.Decode(fixture.Assessments)
	if err != nil {
		t.Fatalf("decode assessments: %v", err)
	}
	item := fixture.Lecture.Data
	lecture := udemy.LecturePlan{
		ID:    fixture.Lecture.ID,
		Title: fixture.Lecture.LectureTitle,
		Quiz:  &udemy.QuizInfo{Title: item.Title, Description: item.Description, PassPercent: item.PassPercent},
	}
	return lecture, decoded.([]pyjson.Value) //nolint:forcetypeassert // reason: every fixture's assessments is a JSON array
}

func assertPythonQuizError(t *testing.T, fixtures fs.FS, stem string, renderErr error) {
	t.Helper()
	recorded, err := fs.ReadFile(fixtures, stem+quizErrorSuffix)
	if err != nil {
		t.Fatalf("no %s or %s golden: %v", quizPageSuffix, quizErrorSuffix, err)
	}
	pythonError := strings.TrimSpace(string(recorded))
	want, known := pythonQuizErrors[pythonError]
	if !known {
		t.Fatalf("no Go error mapped to Python's %s", pythonError)
	}
	if !errors.Is(renderErr, want) {
		t.Errorf("quizPage error = %v, want %v like Python's %s", renderErr, want, pythonError)
	}
}
