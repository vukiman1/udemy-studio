package httpapi

import (
	"context"
	"errors"

	"github.com/vukiman1/udemy-studio/internal/transcode"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func (server apiServer) PlanCourseCompression(ctx context.Context, request gen.PlanCourseCompressionRequestObject) (gen.PlanCourseCompressionResponseObject, error) {
	if server.deps.Compression == nil {
		return gen.PlanCourseCompression503ApplicationProblemPlusJSONResponse{CompressionUnavailableApplicationProblemPlusJSONResponse: compressionUnavailable()}, nil
	}
	plan, err := server.deps.Compression.Plan(ctx, request.CourseId)
	if errors.Is(err, transcode.ErrCourseNotFound) {
		return gen.PlanCourseCompression404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound(err.Error())}, nil
	}
	if err != nil {
		return nil, err
	}
	response := gen.PlanCourseCompression200JSONResponse{
		Chapters:    toCompressionChapters(plan.Chapters()),
		AlreadyHevc: plan.AlreadyHEVC,
		OtherCodec:  plan.OtherCodec,
		NotWorthIt:  plan.NotWorthIt,
		Unreadable:  plan.Unreadable,
	}
	return response, nil
}

func (server apiServer) GetCourseCompression(_ context.Context, request gen.GetCourseCompressionRequestObject) (gen.GetCourseCompressionResponseObject, error) {
	if server.deps.Compression == nil {
		return gen.GetCourseCompression503ApplicationProblemPlusJSONResponse{CompressionUnavailableApplicationProblemPlusJSONResponse: compressionUnavailable()}, nil
	}
	run, ok := server.deps.Compression.Status(request.CourseId)
	if !ok {
		return gen.GetCourseCompression204Response{}, nil
	}
	return gen.GetCourseCompression200JSONResponse(toCompressionRun(run)), nil
}

func (server apiServer) StartCourseCompression(ctx context.Context, request gen.StartCourseCompressionRequestObject) (gen.StartCourseCompressionResponseObject, error) {
	if server.deps.Compression == nil {
		return gen.StartCourseCompression503ApplicationProblemPlusJSONResponse{CompressionUnavailableApplicationProblemPlusJSONResponse: compressionUnavailable()}, nil
	}
	run, err := server.deps.Compression.Start(ctx, request.CourseId, request.Body.Chapters)
	switch {
	case errors.Is(err, transcode.ErrNoChapters), errors.Is(err, transcode.ErrUnknownChapter):
		return gen.StartCourseCompression400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	case errors.Is(err, transcode.ErrCourseNotFound):
		return gen.StartCourseCompression404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound(err.Error())}, nil
	case errors.Is(err, transcode.ErrAlreadyQueued):
		return gen.StartCourseCompression409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict(err)}, nil
	case err != nil:
		return nil, err
	}
	return gen.StartCourseCompression202JSONResponse(toCompressionRun(run)), nil
}

func (server apiServer) CancelCourseCompression(_ context.Context, request gen.CancelCourseCompressionRequestObject) (gen.CancelCourseCompressionResponseObject, error) {
	if server.deps.Compression == nil {
		return gen.CancelCourseCompression503ApplicationProblemPlusJSONResponse{CompressionUnavailableApplicationProblemPlusJSONResponse: compressionUnavailable()}, nil
	}
	err := server.deps.Compression.Cancel(request.CourseId)
	if errors.Is(err, transcode.ErrNotQueued) {
		return gen.CancelCourseCompression409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict(err)}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.CancelCourseCompression204Response{}, nil
}

func toCompressionChapters(chapters []transcode.ChapterPlan) []gen.CompressionChapter {
	converted := make([]gen.CompressionChapter, len(chapters))
	for position, chapter := range chapters {
		converted[position] = gen.CompressionChapter{
			Dir:              chapter.Dir,
			Videos:           chapter.Videos,
			TotalBytes:       chapter.TotalBytes,
			EstimatedSeconds: chapter.EstimatedFor.Seconds(),
		}
	}
	return converted
}

func toCompressionRun(run transcode.Run) gen.CompressionRun {
	converted := gen.CompressionRun{
		State:       gen.CompressionRunState(run.State),
		Total:       run.Total,
		Finished:    run.Finished,
		BytesBefore: run.BytesBefore,
		BytesAfter:  run.BytesAfter,
		Skipped:     toCompressionOutcomes(run.Skipped),
		Failed:      toCompressionOutcomes(run.Failed),
	}
	if run.Current != nil {
		converted.Current = &gen.CompressionCurrent{
			LectureTitle:   run.Current.LectureTitle,
			EncodedSeconds: run.Current.Encoded.Seconds(),
			LengthSeconds:  run.Current.Length.Seconds(),
		}
	}
	return converted
}

func toCompressionOutcomes(outcomes []transcode.Outcome) []gen.CompressionOutcome {
	converted := make([]gen.CompressionOutcome, len(outcomes))
	for position, outcome := range outcomes {
		converted[position] = gen.CompressionOutcome{Path: outcome.Path, LectureTitle: outcome.LectureTitle, Reason: outcome.Reason}
	}
	return converted
}
