package learning

import (
	"slices"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/progress"
)

type Target struct {
	CourseID        string
	LectureID       string
	PositionSeconds float64
}

type CourseOverview struct {
	Course            library.Course
	PlayableLectures  int
	CompletedLectures int
	Resume            *Target
}

func Overviews(scanned *library.Library, records map[string]progress.Record) []CourseOverview {
	overviews := make([]CourseOverview, 0, len(scanned.Courses))
	for _, course := range scanned.Courses {
		overviews = append(overviews, Overview(course, records))
	}
	return overviews
}

func Overview(course library.Course, records map[string]progress.Record) CourseOverview {
	overview := CourseOverview{Course: course}
	for _, lecture := range playableLectures(course) {
		overview.PlayableLectures++
		if records[lecture.Key].IsCompleted() {
			overview.CompletedLectures++
		}
	}
	if resume, ok := resumeCourse(course, records); ok {
		overview.Resume = &resume.target
	}
	return overview
}

func Resume(scanned *library.Library, records map[string]progress.Record) (Target, bool) {
	var best courseResume
	found := false
	for _, course := range scanned.Courses {
		candidate, ok := resumeCourse(course, records)
		if ok && (!found || candidate.updatedAt.After(best.updatedAt)) {
			best, found = candidate, true
		}
	}
	return best.target, found
}

type courseResume struct {
	target    Target
	updatedAt time.Time
}

func resumeCourse(course library.Course, records map[string]progress.Record) (courseResume, bool) {
	lectures := playableLectures(course)
	latest := latestRecord(lectures, records)
	position := slices.IndexFunc(lectures, func(lecture library.Lecture) bool { return lecture.Key == latest.LectureKey })
	switch {
	case position < 0:
		return courseResume{}, false
	case !latest.IsCompleted():
		target := Target{CourseID: course.ID, LectureID: lectures[position].ID, PositionSeconds: latest.PositionSeconds}
		return courseResume{target: target, updatedAt: latest.UpdatedAt}, true
	case position+1 < len(lectures):
		target := Target{CourseID: course.ID, LectureID: lectures[position+1].ID}
		return courseResume{target: target, updatedAt: latest.UpdatedAt}, true
	default:
		return courseResume{}, false
	}
}

func latestRecord(lectures []library.Lecture, records map[string]progress.Record) progress.Record {
	var latest progress.Record
	for _, lecture := range lectures {
		record, ok := records[lecture.Key]
		if ok && record.UpdatedAt.After(latest.UpdatedAt) {
			latest = record
		}
	}
	return latest
}

func playableLectures(course library.Course) []library.Lecture {
	var lectures []library.Lecture
	for _, chapter := range course.Chapters {
		for _, lecture := range chapter.Lectures {
			if lecture.Kind != library.LectureKindUnavailable {
				lectures = append(lectures, lecture)
			}
		}
	}
	return lectures
}
