package udemyapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vukiman1/udemy-studio/internal/platform/pyjson"
)

const quizAssessmentsPathTemplate = "/api-2.0/quizzes/%d/assessments/?page_size=250&fields[assessment]=id,assessment_type,prompt,correct_response,section,question_plain,related_lectures"

// QuizAssessments returns every question of a quiz with its keys and numbers as the API sent them.
func (client *Client) QuizAssessments(ctx context.Context, quizID int64) ([]pyjson.Value, error) {
	raw, err := allPages[json.RawMessage](ctx, client, client.portalURL+fmt.Sprintf(quizAssessmentsPathTemplate, quizID))
	if err != nil {
		return nil, fmt.Errorf("fetch quiz %d: %w", quizID, err)
	}
	assessments := make([]pyjson.Value, 0, len(raw))
	for _, item := range raw {
		assessment, err := pyjson.Decode(item)
		if err != nil {
			return nil, fmt.Errorf("decode an assessment of quiz %d: %w", quizID, err)
		}
		assessments = append(assessments, assessment)
	}
	return assessments, nil
}
