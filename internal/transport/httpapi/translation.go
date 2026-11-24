package httpapi

import (
	"context"
	"errors"

	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/translate"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func (server apiServer) PlanCourseTranslation(ctx context.Context, request gen.PlanCourseTranslationRequestObject) (gen.PlanCourseTranslationResponseObject, error) {
	plan, err := server.deps.Translation.Plan(ctx, request.CourseId)
	switch {
	case errors.Is(err, llm.ErrNotConfigured):
		return gen.PlanCourseTranslation400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	case errors.Is(err, translate.ErrCourseNotFound):
		return gen.PlanCourseTranslation404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound(err.Error())}, nil
	case err != nil:
		return nil, err
	}
	return gen.PlanCourseTranslation200JSONResponse{
		Lectures:          len(plan.Candidates),
		AlreadyTranslated: plan.AlreadyTranslated,
		Characters:        plan.Characters,
		EstimatedUsd:      plan.EstimatedUSD,
		IsPriced:          plan.IsPriced,
	}, nil
}

func (server apiServer) GetCourseTranslation(_ context.Context, request gen.GetCourseTranslationRequestObject) (gen.GetCourseTranslationResponseObject, error) {
	run, ok := server.deps.Translation.Status(request.CourseId)
	if !ok {
		return gen.GetCourseTranslation204Response{}, nil
	}
	return gen.GetCourseTranslation200JSONResponse(toTranslationRun(run)), nil
}

func (server apiServer) StartCourseTranslation(ctx context.Context, request gen.StartCourseTranslationRequestObject) (gen.StartCourseTranslationResponseObject, error) {
	run, err := server.deps.Translation.Start(ctx, request.CourseId)
	switch {
	case errors.Is(err, llm.ErrNotConfigured):
		return gen.StartCourseTranslation400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	case errors.Is(err, translate.ErrCourseNotFound):
		return gen.StartCourseTranslation404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound(err.Error())}, nil
	case errors.Is(err, translate.ErrAlreadyQueued):
		return gen.StartCourseTranslation409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict(err)}, nil
	case err != nil:
		return nil, err
	}
	return gen.StartCourseTranslation202JSONResponse(toTranslationRun(run)), nil
}

func (server apiServer) CancelCourseTranslation(_ context.Context, request gen.CancelCourseTranslationRequestObject) (gen.CancelCourseTranslationResponseObject, error) {
	err := server.deps.Translation.Cancel(request.CourseId)
	if errors.Is(err, translate.ErrNotQueued) {
		return gen.CancelCourseTranslation409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict(err)}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.CancelCourseTranslation204Response{}, nil
}

func toTranslationRun(run translate.Run) gen.TranslationRun {
	converted := gen.TranslationRun{
		State:    gen.TranslationRunState(run.State),
		Total:    run.Total,
		Finished: run.Finished,
		CostUsd:  run.CostUSD,
		Failed:   make([]gen.TranslationOutcome, len(run.Failed)),
	}
	for position, outcome := range run.Failed {
		converted.Failed[position] = gen.TranslationOutcome{Path: outcome.Path, LectureTitle: outcome.LectureTitle, Reason: outcome.Reason}
	}
	if run.CurrentLecture != "" {
		converted.CurrentLecture = &run.CurrentLecture
	}
	if run.Error != "" {
		converted.Error = &run.Error
	}
	return converted
}
