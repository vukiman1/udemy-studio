package httpapi

import (
	"maps"
	"net/http"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func TestListCoursesCountsPlayableAndCompletedLectures(t *testing.T) {
	app := newTestApp(t)
	courses := decode[[]gen.CourseSummary](t, app.do(t, http.MethodGet, "/api/courses", nil))
	if len(courses) != 1 || courses[0].PlayableLectures != 3 || courses[0].CompletedLectures != 0 || courses[0].Resume != nil {
		t.Fatalf("courses = %+v, want one untouched course with 3 playable lectures", courses)
	}

	app.do(t, http.MethodPut, "/api/lectures/"+app.lectureID(t, 1)+"/progress", gen.ProgressReport{PositionSeconds: 3, DurationSeconds: 3, Completed: true})
	course := decode[[]gen.CourseSummary](t, app.do(t, http.MethodGet, "/api/courses", nil))[0]
	if course.CompletedLectures != 1 || course.Resume == nil || course.Resume.LectureId != app.lectureID(t, 2) {
		t.Errorf("course = %+v, want 1 completed and resume at lecture 002", course)
	}
}

func TestGetCourseDescribesEachLecture(t *testing.T) {
	app := newTestApp(t)
	courseID := app.library.Courses[0].ID
	detail := decode[gen.CourseDetail](t, app.do(t, http.MethodGet, "/api/courses/"+courseID, nil))
	if len(detail.Chapters) != 1 || len(detail.Chapters[0].Lectures) != 4 {
		t.Fatalf("detail = %+v, want 1 chapter with 4 lectures", detail)
	}
	lectures := detail.Chapters[0].Lectures
	video, article, drm := lectures[0], lectures[1], lectures[2]
	if video.Kind != gen.Video || !video.HasVideo || len(video.SubtitleLanguages) != 2 {
		t.Errorf("video lecture = %+v", video)
	}
	if article.Kind != gen.Article || !article.HasArticle || len(article.Attachments) != 1 || article.Attachments[0].Name != "slides.pdf" {
		t.Errorf("article lecture = %+v", article)
	}
	if drm.Kind != gen.Unavailable || drm.HasVideo {
		t.Errorf("drm lecture = %+v", drm)
	}
}

func TestGetUnknownCourseIsAProblem(t *testing.T) {
	assertProblem(t, newTestApp(t).do(t, http.MethodGet, "/api/courses/nope", nil), http.StatusNotFound)
}

func TestContinueIsEmptyUntilSomethingIsWatched(t *testing.T) {
	app := newTestApp(t)
	if result := app.do(t, http.MethodGet, "/api/continue", nil); result.status != http.StatusNoContent {
		t.Fatalf("status = %d, want 204 before any progress", result.status)
	}
	app.do(t, http.MethodPut, "/api/lectures/"+app.lectureID(t, 1)+"/progress", gen.ProgressReport{PositionSeconds: 1.25, DurationSeconds: 3})
	target := decode[gen.ContinueTarget](t, app.do(t, http.MethodGet, "/api/continue", nil))
	if target.LectureId != app.lectureID(t, 1) || target.PositionSeconds != 1.25 {
		t.Errorf("target = %+v, want lecture 001 at 1.25s", target)
	}
}

func TestACourseCoverIsListedAndServedOnlyWhenDownloaded(t *testing.T) {
	withCover := maps.Clone(libraryFixture)
	withCover["course/cover.jpg"] = "jpeg bytes"
	app := newTestAppWith(t, withCover)
	course := decode[[]gen.CourseSummary](t, app.do(t, http.MethodGet, "/api/courses", nil))[0]
	if course.CoverUrl == nil {
		t.Fatal("coverUrl missing for a course with cover.jpg")
	}
	cover := app.do(t, http.MethodGet, *course.CoverUrl, nil)
	if cover.status != http.StatusOK || cover.header.Get("Content-Type") != "image/jpeg" || string(cover.body) != "jpeg bytes" {
		t.Errorf("cover = %d %q %q", cover.status, cover.header.Get("Content-Type"), cover.body)
	}
	plain := newTestApp(t)
	without := decode[[]gen.CourseSummary](t, plain.do(t, http.MethodGet, "/api/courses", nil))[0]
	if without.CoverUrl != nil {
		t.Errorf("coverUrl = %q for a course without a cover", *without.CoverUrl)
	}
	assertProblem(t, plain.do(t, http.MethodGet, "/covers/"+without.Id, nil), http.StatusNotFound)
}
