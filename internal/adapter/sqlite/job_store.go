package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite/gen"
	"github.com/vukiman1/udemy-studio/internal/jobs"
)

type JobStore struct {
	queries *gen.Queries
}

func NewJobStore(database *sql.DB) *JobStore {
	return &JobStore{queries: gen.New(database)}
}

type storedJob struct {
	ID             string     `json:"id"`
	CourseURL      string     `json:"course_url"`
	Chapters       string     `json:"chapters"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	CourseTitle    string     `json:"course_title"`
	TotalLectures  int        `json:"total_lectures"`
	FinishedCount  int        `json:"finished_count"`
	CurrentLecture string     `json:"current_lecture"`
	Failures       []string   `json:"failures"`
	Error          string     `json:"error"`
	ErrorKind      string     `json:"error_kind,omitempty"`
}

func (store *JobStore) Save(ctx context.Context, job jobs.Job) error {
	state, err := json.Marshal(storedJob{
		ID: job.ID, CourseURL: job.Request.CourseURL, Chapters: job.Request.Chapters, Status: string(job.Status),
		CreatedAt: job.CreatedAt, StartedAt: job.StartedAt, FinishedAt: job.FinishedAt,
		CourseTitle: job.Progress.CourseTitle, TotalLectures: job.Progress.TotalLectures,
		FinishedCount: job.Progress.FinishedCount, CurrentLecture: job.Progress.CurrentLecture,
		Failures: job.Failures, Error: job.Error, ErrorKind: string(job.ErrorKind),
	})
	if err != nil {
		return fmt.Errorf("encode job %s: %w", job.ID, err)
	}
	if err := store.queries.UpsertJob(ctx, gen.UpsertJobParams{
		ID: job.ID, Status: string(job.Status), CreatedAt: formatTime(job.CreatedAt), StateJson: string(state),
	}); err != nil {
		return fmt.Errorf("save job %s: %w", job.ID, err)
	}
	return nil
}

func (store *JobStore) List(ctx context.Context) ([]jobs.Job, error) {
	states, err := store.queries.ListJobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	list := make([]jobs.Job, 0, len(states))
	for _, state := range states {
		var stored storedJob
		if err := json.Unmarshal([]byte(state), &stored); err != nil {
			return nil, fmt.Errorf("decode stored job: %w", err)
		}
		list = append(list, jobs.Job{
			ID:         stored.ID,
			Request:    jobs.Request{CourseURL: stored.CourseURL, Chapters: stored.Chapters},
			Status:     jobs.Status(stored.Status),
			CreatedAt:  stored.CreatedAt,
			StartedAt:  stored.StartedAt,
			FinishedAt: stored.FinishedAt,
			Progress: jobs.Progress{
				CourseTitle: stored.CourseTitle, TotalLectures: stored.TotalLectures,
				FinishedCount: stored.FinishedCount, CurrentLecture: stored.CurrentLecture,
			},
			Failures:  stored.Failures,
			Error:     stored.Error,
			ErrorKind: jobs.ErrorKind(stored.ErrorKind),
		})
	}
	return list, nil
}
