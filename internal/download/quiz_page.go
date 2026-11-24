package download

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vukiman1/udemy-studio/internal/platform/pyjson"
	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const (
	assessmentTypeKey    = "assessment_type"
	classKey             = "_class"
	promptKey            = "prompt"
	instructionsKey      = "instructions"
	testFilesKey         = "test_files"
	solutionFilesKey     = "solution_files"
	quizTitlePlaceholder = "%%TITLE%%"
	quizJSONPlaceholder  = "%%QUIZ_JSON%%"
	assessmentClass      = "assessment"
	codingProblemType    = "coding-problem"
	missingPromptPart    = "(None)"
)

var (
	//go:embed templates/quiz_template.html
	quizTemplate string
	//go:embed templates/coding_assignment_template.html
	codingAssignmentTemplate string

	errEmptyQuiz          = errors.New("quiz has no assessments")
	errMalformedQuiz      = errors.New("quiz's first assessment is not an object")
	errMissingQuizField   = errors.New("quiz's first assessment lacks a field the page needs")
	pythonHTMLEscapes     = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#x27;")
	scriptCloseTagEscapes = strings.NewReplacer("</", `<\/`)
)

func (run *courseDownload) quiz(ctx context.Context, chapterDir string, lecture udemy.LecturePlan) error {
	failure := fmt.Sprintf("%s (%s)", lecture.Title, lecture.Class)
	assessments, err := run.tools.Quizzes.QuizAssessments(ctx, lecture.ID)
	if err != nil {
		return run.recordFailure(ctx, failure, err)
	}
	page, err := quizPage(lecture, assessments)
	if err == nil {
		err = writeLibraryFile(filepath.Join(chapterDir, lecture.File), []byte(page))
	}
	if err != nil {
		return run.recordFailure(ctx, failure, err)
	}
	run.summary.Downloaded++
	return nil
}

// quizPage renders a quiz like the Python downloader: a lone coding problem becomes a coding assignment page.
func quizPage(lecture udemy.LecturePlan, assessments []pyjson.Value) (string, error) {
	if len(assessments) == 0 {
		return "", errEmptyQuiz
	}
	first, ok := assessments[0].(*pyjson.Object)
	if !ok {
		return "", errMalformedQuiz
	}
	assessmentType, hasType := first.Get(assessmentTypeKey)
	class, hasClass := first.Get(classKey)
	if !hasType {
		return "", fmt.Errorf("%w: %s", errMissingQuizField, assessmentTypeKey)
	}
	if !hasClass {
		return "", fmt.Errorf("%w: %s", errMissingQuizField, classKey)
	}
	if len(assessments) == 1 && class == assessmentClass && assessmentType == codingProblemType {
		return codingAssignmentPage(lecture, first)
	}
	return normalQuizPage(lecture, assessments)
}

func normalQuizPage(lecture udemy.LecturePlan, assessments []pyjson.Value) (string, error) {
	description, err := rawValue(lecture.Quiz.Description)
	if err != nil {
		return "", err
	}
	passPercent, err := rawValue(lecture.Quiz.PassPercent)
	if err != nil {
		return "", err
	}
	data := pyjson.NewObject().
		Set("id", lecture.ID).
		Set("title", lecture.Quiz.Title).
		Set("description", description).
		Set("pass_score", passPercent).
		Set("assessments", assessments)
	embedded, err := jsonForScript(data)
	if err != nil {
		return "", err
	}
	page := strings.ReplaceAll(quizTemplate, quizTitlePlaceholder, pythonHTMLEscapes.Replace(lecture.Quiz.Title))
	return strings.ReplaceAll(page, quizJSONPlaceholder, embedded), nil
}

func codingAssignmentPage(lecture udemy.LecturePlan, assignment *pyjson.Object) (string, error) {
	prompt, ok := assignment.Get(promptKey)
	if !ok {
		return "", fmt.Errorf("%w: %s", errMissingQuizField, promptKey)
	}
	instructions, tests, solutions := promptPart(prompt, instructionsKey), promptPart(prompt, testFilesKey), promptPart(prompt, solutionFilesKey)
	_, testsMissing := tests.(string)
	_, solutionsMissing := solutions.(string)
	data := pyjson.NewObject().
		Set("title", lecture.Title).
		Set("hasInstructions", instructions != missingPromptPart).
		Set("hasTests", !testsMissing).
		Set("hasSolutions", !solutionsMissing).
		Set("instructions", instructions).
		Set("tests", tests).
		Set("solutions", solutions)
	embedded, err := jsonForScript(data)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(codingAssignmentTemplate, dataPlaceholder, embedded), nil
}

// promptPart mirrors the Python downloader's lookup, which stands in "(None)" for a missing part.
func promptPart(prompt pyjson.Value, key string) pyjson.Value {
	object, ok := prompt.(*pyjson.Object)
	if !ok {
		return missingPromptPart
	}
	if value, ok := object.Get(key); ok {
		return value
	}
	return missingPromptPart
}

// jsonForScript keeps a "</script>" inside the data from closing the page's script tag early.
func jsonForScript(value pyjson.Value) (string, error) {
	encoded, err := pyjson.Dumps(value)
	if err != nil {
		return "", fmt.Errorf("encode quiz data: %w", err)
	}
	return scriptCloseTagEscapes.Replace(encoded), nil
}

// rawValue reads a curriculum field the API may leave out, which Python then embeds as null.
func rawValue(raw []byte) (pyjson.Value, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	value, err := pyjson.Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("decode quiz field: %w", err)
	}
	return value, nil
}
