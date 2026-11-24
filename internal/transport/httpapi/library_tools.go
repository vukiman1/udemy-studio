package httpapi

import (
	"context"

	"github.com/vukiman1/udemy-studio/internal/translate"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func (server apiServer) ListCourseTools(ctx context.Context, _ gen.ListCourseToolsRequestObject) (gen.ListCourseToolsResponseObject, error) {
	courses := server.deps.Catalog.Current().Courses
	response := make(gen.ListCourseTools200JSONResponse, len(courses))
	for position, course := range courses {
		pending, alreadyTranslated := translate.PendingLectures(course)
		tools := gen.CourseTools{
			CourseId:    course.ID,
			Translation: gen.TranslationTools{Lectures: len(pending) + alreadyTranslated, Pending: len(pending)},
		}
		if run, ok := server.deps.Translation.Status(course.ID); ok {
			converted := toTranslationRun(run)
			tools.Translation.Run = &converted
		}
		if server.deps.Compression != nil {
			summary, err := server.deps.Compression.Summarize(ctx, course)
			if err != nil {
				return nil, err
			}
			compression := gen.CompressionTools{Videos: summary.Videos, Pending: summary.Pending, Unchecked: summary.Unchecked}
			if run, ok := server.deps.Compression.Status(course.ID); ok {
				converted := toCompressionRun(run)
				compression.Run = &converted
			}
			tools.Compression = &compression
		}
		response[position] = tools
	}
	return response, nil
}
