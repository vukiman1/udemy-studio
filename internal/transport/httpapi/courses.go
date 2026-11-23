package httpapi

import (
	"context"
	"fmt"

	"github.com/vukiman1/udemy-studio/internal/learning"
	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/progress"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func (server apiServer) ListCourses(ctx context.Context, _ gen.ListCoursesRequestObject) (gen.ListCoursesResponseObject, error) {
	records, err := server.deps.Progress.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}
	overviews := learning.Overviews(server.deps.Library, records)
	summaries := make(gen.ListCourses200JSONResponse, 0, len(overviews))
	for _, overview := range overviews {
		summaries = append(summaries, toCourseSummary(overview))
	}
	return summaries, nil
}

func (server apiServer) GetCourse(ctx context.Context, request gen.GetCourseRequestObject) (gen.GetCourseResponseObject, error) {
	if _, ok := server.deps.Library.Course(request.CourseId); !ok {
		return gen.GetCourse404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound("no course " + request.CourseId)}, nil
	}
	records, err := server.deps.Progress.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get course %s: %w", request.CourseId, err)
	}
	for _, overview := range learning.Overviews(server.deps.Library, records) {
		if overview.Course.ID == request.CourseId {
			return gen.GetCourse200JSONResponse(toCourseDetail(overview, records)), nil
		}
	}
	return nil, fmt.Errorf("course %s vanished between lookup and overview", request.CourseId)
}

func (server apiServer) GetContinue(ctx context.Context, _ gen.GetContinueRequestObject) (gen.GetContinueResponseObject, error) {
	records, err := server.deps.Progress.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get continue target: %w", err)
	}
	target, ok := learning.Resume(server.deps.Library, records)
	if !ok {
		return gen.GetContinue204Response{}, nil
	}
	return gen.GetContinue200JSONResponse(toContinueTarget(target)), nil
}

func toCourseSummary(overview learning.CourseOverview) gen.CourseSummary {
	summary := gen.CourseSummary{
		Id:                overview.Course.ID,
		Title:             overview.Course.Title,
		PlayableLectures:  overview.PlayableLectures,
		CompletedLectures: overview.CompletedLectures,
	}
	if overview.Resume != nil {
		target := toContinueTarget(*overview.Resume)
		summary.Resume = &target
	}
	return summary
}

func toCourseDetail(overview learning.CourseOverview, records map[string]progress.Record) gen.CourseDetail {
	summary := toCourseSummary(overview)
	detail := gen.CourseDetail{
		Id:                summary.Id,
		Title:             summary.Title,
		PlayableLectures:  summary.PlayableLectures,
		CompletedLectures: summary.CompletedLectures,
		Resume:            summary.Resume,
		Chapters:          make([]gen.Chapter, 0, len(overview.Course.Chapters)),
	}
	for _, chapter := range overview.Course.Chapters {
		lectures := make([]gen.LectureSummary, 0, len(chapter.Lectures))
		for _, lecture := range chapter.Lectures {
			lectures = append(lectures, toLectureSummary(lecture, records))
		}
		detail.Chapters = append(detail.Chapters, gen.Chapter{Index: chapter.Index, Title: chapter.Title, Lectures: lectures})
	}
	return detail
}

func toLectureSummary(lecture library.Lecture, records map[string]progress.Record) gen.LectureSummary {
	summary := gen.LectureSummary{
		Id:                lecture.ID,
		Number:            lecture.Number,
		Title:             lecture.Title,
		Kind:              gen.LectureKind(lecture.Kind),
		HasVideo:          lecture.VideoPath != "",
		HasArticle:        lecture.ArticlePath != "",
		SubtitleLanguages: make([]string, 0, len(lecture.Subtitles)),
		Attachments:       make([]gen.Attachment, 0, len(lecture.Attachments)),
	}
	for _, subtitle := range lecture.Subtitles {
		summary.SubtitleLanguages = append(summary.SubtitleLanguages, subtitle.Language)
	}
	for _, attachment := range lecture.Attachments {
		summary.Attachments = append(summary.Attachments, gen.Attachment{Id: attachment.ID, Name: attachment.Name})
	}
	if record, ok := records[lecture.Key]; ok {
		lectureProgress := toLectureProgress(record)
		summary.Progress = &lectureProgress
	}
	return summary
}

func toLectureProgress(record progress.Record) gen.LectureProgress {
	return gen.LectureProgress{
		PositionSeconds: record.PositionSeconds,
		DurationSeconds: record.DurationSeconds,
		Completed:       record.IsCompleted(),
		UpdatedAt:       record.UpdatedAt,
	}
}

func toContinueTarget(target learning.Target) gen.ContinueTarget {
	return gen.ContinueTarget{CourseId: target.CourseID, LectureId: target.LectureID, PositionSeconds: target.PositionSeconds}
}
