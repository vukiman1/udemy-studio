package httpapi

import (
	"context"
	"errors"

	"github.com/vukiman1/udemy-studio/internal/objectstore"
	"github.com/vukiman1/udemy-studio/internal/publish"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func (server apiServer) PlanCoursePublish(ctx context.Context, request gen.PlanCoursePublishRequestObject) (gen.PlanCoursePublishResponseObject, error) {
	plan, err := server.deps.Publish.Plan(ctx, request.CourseId)
	switch {
	case errors.Is(err, publish.ErrCourseNotFound):
		return gen.PlanCoursePublish404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound(err.Error())}, nil
	case errors.Is(err, objectstore.ErrNotConfigured):
		return gen.PlanCoursePublish409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict(err)}, nil
	case isBadStorageRequest(err):
		return gen.PlanCoursePublish400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	case errors.Is(err, objectstore.ErrUnreachable):
		return gen.PlanCoursePublish502ApplicationProblemPlusJSONResponse{ProviderUnreachableApplicationProblemPlusJSONResponse: providerUnreachable(err)}, nil
	case err != nil:
		return nil, err
	}
	response := gen.PlanCoursePublish200JSONResponse{
		SharedBytes:   plan.SharedBytes(),
		LimitBytes:    plan.LimitBytes,
		Chapters:      toPublishChapters(plan.Chapters()),
		Replaces:      plan.Replaces(),
		SharedChanged: plan.SharedChanged(),
	}
	if plan.Published != nil {
		published := server.publishedCourse(*plan.Published)
		response.Published = &published
	}
	return response, nil
}

func (server apiServer) StartCoursePublish(ctx context.Context, request gen.StartCoursePublishRequestObject) (gen.StartCoursePublishResponseObject, error) {
	run, err := server.deps.Publish.Start(ctx, request.CourseId, request.Body.Chapters)
	switch {
	case err == nil:
		return gen.StartCoursePublish202JSONResponse(toPublishRun(run)), nil
	case errors.Is(err, publish.ErrCourseNotFound):
		return gen.StartCoursePublish404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound(err.Error())}, nil
	case errors.Is(err, objectstore.ErrNotConfigured), errors.Is(err, publish.ErrBusy), errors.Is(err, publish.ErrTooLarge):
		return gen.StartCoursePublish409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict(err)}, nil
	case isBadStorageRequest(err), errors.Is(err, publish.ErrNoChapters), errors.Is(err, publish.ErrUnknownChapter):
		return gen.StartCoursePublish400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	case errors.Is(err, objectstore.ErrUnreachable):
		return gen.StartCoursePublish502ApplicationProblemPlusJSONResponse{ProviderUnreachableApplicationProblemPlusJSONResponse: providerUnreachable(err)}, nil
	default:
		return nil, err
	}
}

func (server apiServer) GetPublishRun(_ context.Context, _ gen.GetPublishRunRequestObject) (gen.GetPublishRunResponseObject, error) {
	run, found := server.deps.Publish.Status()
	if !found {
		return gen.GetPublishRun204Response{}, nil
	}
	return gen.GetPublishRun200JSONResponse(toPublishRun(run)), nil
}

func (server apiServer) CancelPublish(_ context.Context, _ gen.CancelPublishRequestObject) (gen.CancelPublishResponseObject, error) {
	err := server.deps.Publish.Cancel()
	if errors.Is(err, publish.ErrNotRunning) {
		return gen.CancelPublish409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict(err)}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.CancelPublish204Response{}, nil
}

func (server apiServer) GetPublishedCourse(ctx context.Context, _ gen.GetPublishedCourseRequestObject) (gen.GetPublishedCourseResponseObject, error) {
	index, found, err := server.deps.Publish.Published(ctx)
	switch {
	case errors.Is(err, objectstore.ErrNotConfigured):
		return gen.GetPublishedCourse204Response{}, nil
	case err != nil:
		return gen.GetPublishedCourse502ApplicationProblemPlusJSONResponse{ProviderUnreachableApplicationProblemPlusJSONResponse: providerUnreachable(err)}, nil
	case !found:
		return gen.GetPublishedCourse204Response{}, nil
	}
	return gen.GetPublishedCourse200JSONResponse(server.publishedCourse(index)), nil
}

func (server apiServer) RemovePublishedCourse(ctx context.Context, _ gen.RemovePublishedCourseRequestObject) (gen.RemovePublishedCourseResponseObject, error) {
	err := server.deps.Publish.Remove(ctx)
	switch {
	case err == nil:
		return gen.RemovePublishedCourse204Response{}, nil
	case errors.Is(err, objectstore.ErrNotConfigured), errors.Is(err, publish.ErrBusy):
		return gen.RemovePublishedCourse409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict(err)}, nil
	case isBadStorageRequest(err):
		return gen.RemovePublishedCourse400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	case errors.Is(err, objectstore.ErrUnreachable):
		return gen.RemovePublishedCourse502ApplicationProblemPlusJSONResponse{ProviderUnreachableApplicationProblemPlusJSONResponse: providerUnreachable(err)}, nil
	default:
		return nil, err
	}
}

func (server apiServer) publishedCourse(index publish.Index) gen.PublishedCourse {
	published := gen.PublishedCourse{
		Dir:           index.Course.Dir,
		Title:         index.Course.Title,
		SizeBytes:     index.Course.SizeBytes,
		PublishedAt:   index.PublishedAt,
		Chapters:      index.PublishedChapters(),
		TotalChapters: index.Course.ChapterCount,
	}
	for _, course := range server.deps.Catalog.Current().Courses {
		if course.Dir == index.Course.Dir {
			published.CourseId = &course.ID
			break
		}
	}
	return published
}

func toPublishRun(run publish.Run) gen.PublishRun {
	converted := gen.PublishRun{
		CourseId:     run.CourseID,
		Title:        run.Title,
		State:        gen.PublishRunState(run.State),
		TotalFiles:   run.TotalFiles,
		SentFiles:    run.SentFiles,
		SkippedFiles: run.SkippedFiles,
		TotalBytes:   run.TotalBytes,
		SentBytes:    run.SentBytes,
	}
	if run.CurrentFile != "" {
		converted.CurrentFile = &run.CurrentFile
	}
	if run.Failure != "" {
		converted.Failure = &run.Failure
	}
	return converted
}

func toPublishChapters(chapters []publish.ChapterPlan) []gen.PublishChapter {
	converted := make([]gen.PublishChapter, len(chapters))
	for position, chapter := range chapters {
		converted[position] = gen.PublishChapter{Dir: chapter.Dir, Files: chapter.Files, SizeBytes: chapter.SizeBytes, Published: chapter.IsPublished, Changed: chapter.HasChanges}
	}
	return converted
}
