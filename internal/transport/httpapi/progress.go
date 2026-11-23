package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/vukiman1/udemy-studio/internal/progress"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func (server apiServer) PutLectureProgress(ctx context.Context, request gen.PutLectureProgressRequestObject) (gen.PutLectureProgressResponseObject, error) {
	lecture, ok := server.deps.Catalog.Current().Lecture(request.LectureId)
	if !ok {
		return gen.PutLectureProgress404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound("no lecture " + request.LectureId)}, nil
	}
	record, err := server.deps.Progress.Save(ctx, progress.Report{
		LectureKey:      lecture.Key,
		PositionSeconds: request.Body.PositionSeconds,
		DurationSeconds: request.Body.DurationSeconds,
		Completed:       request.Body.Completed,
	})
	if errors.Is(err, progress.ErrInvalidReport) {
		badRequest := gen.BadRequestApplicationProblemPlusJSONResponse(problem(http.StatusBadRequest, err.Error()))
		return gen.PutLectureProgress400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("save progress of lecture %s: %w", lecture.ID, err)
	}
	return gen.PutLectureProgress200JSONResponse(toLectureProgress(record)), nil
}
