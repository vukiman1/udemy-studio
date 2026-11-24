package httpapi

import (
	"context"

	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func (server apiServer) SearchSubtitles(ctx context.Context, request gen.SearchSubtitlesRequestObject) (gen.SearchSubtitlesResponseObject, error) {
	hits, err := server.deps.Search.Search(ctx, request.Params.Q)
	if err != nil {
		return nil, err
	}
	scanned := server.deps.Catalog.Current()
	results := make(gen.SearchSubtitles200JSONResponse, 0, len(hits))
	for _, hit := range hits {
		course, lecture, ok := scanned.LectureByKey(hit.LectureKey)
		if !ok {
			continue
		}
		results = append(results, gen.SearchHit{
			CourseId:     course.ID,
			CourseTitle:  course.Title,
			LectureId:    lecture.ID,
			LectureTitle: lecture.Title,
			Language:     hit.Language,
			StartMs:      hit.Start.Milliseconds(),
			Snippet:      hit.Snippet,
		})
	}
	return results, nil
}
