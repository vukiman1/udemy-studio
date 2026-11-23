package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func (server apiServer) ListJobs(_ context.Context, _ gen.ListJobsRequestObject) (gen.ListJobsResponseObject, error) {
	if server.deps.Jobs == nil {
		return gen.ListJobs503ApplicationProblemPlusJSONResponse{DownloadsUnavailableApplicationProblemPlusJSONResponse: downloadsUnavailable()}, nil
	}
	listed := server.deps.Jobs.List()
	response := make(gen.ListJobs200JSONResponse, 0, len(listed))
	for _, job := range listed {
		response = append(response, toJob(job))
	}
	return response, nil
}

func (server apiServer) CreateJob(ctx context.Context, request gen.CreateJobRequestObject) (gen.CreateJobResponseObject, error) {
	if server.deps.Jobs == nil {
		return gen.CreateJob503ApplicationProblemPlusJSONResponse{DownloadsUnavailableApplicationProblemPlusJSONResponse: downloadsUnavailable()}, nil
	}
	submitted := jobs.Request{CourseURL: request.Body.CourseUrl}
	if request.Body.Chapters != nil {
		submitted.Chapters = *request.Body.Chapters
	}
	job, err := server.deps.Jobs.Submit(ctx, submitted)
	if errors.Is(err, jobs.ErrInvalidRequest) {
		badRequest := gen.BadRequestApplicationProblemPlusJSONResponse(problem(http.StatusBadRequest, err.Error()))
		return gen.CreateJob400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.CreateJob202JSONResponse(toJob(job)), nil
}

func (server apiServer) CancelJob(ctx context.Context, request gen.CancelJobRequestObject) (gen.CancelJobResponseObject, error) {
	if server.deps.Jobs == nil {
		return gen.CancelJob503ApplicationProblemPlusJSONResponse{DownloadsUnavailableApplicationProblemPlusJSONResponse: downloadsUnavailable()}, nil
	}
	err := server.deps.Jobs.Cancel(ctx, request.JobId)
	switch {
	case errors.Is(err, jobs.ErrJobNotFound):
		return gen.CancelJob404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound(err.Error())}, nil
	case errors.Is(err, jobs.ErrJobFinished):
		conflict := gen.ConflictApplicationProblemPlusJSONResponse(problem(http.StatusConflict, err.Error()))
		return gen.CancelJob409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict}, nil
	case err != nil:
		return nil, err
	}
	return gen.CancelJob204Response{}, nil
}

func toJob(job jobs.Job) gen.Job {
	response := gen.Job{
		Id:         job.ID,
		CourseUrl:  job.Request.CourseURL,
		Status:     gen.JobStatus(job.Status),
		CreatedAt:  job.CreatedAt,
		StartedAt:  job.StartedAt,
		FinishedAt: job.FinishedAt,
		Failures:   job.Failures,
		Progress: gen.JobProgress{
			CourseTitle:   job.Progress.CourseTitle,
			TotalLectures: job.Progress.TotalLectures,
			FinishedCount: job.Progress.FinishedCount,
		},
	}
	if response.Failures == nil {
		response.Failures = []string{}
	}
	if job.Request.Chapters != "" {
		response.Chapters = &job.Request.Chapters
	}
	if job.Progress.CurrentLecture != "" {
		response.Progress.CurrentLecture = &job.Progress.CurrentLecture
	}
	if job.Error != "" {
		response.Error = &job.Error
	}
	return response
}
