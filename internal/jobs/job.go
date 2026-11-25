package jobs

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

func (status Status) IsFinished() bool {
	return status == StatusSucceeded || status == StatusFailed || status == StatusCancelled
}

type ErrorKind string

const (
	ErrorKindLogin          ErrorKind = "login"
	ErrorKindCourseNotFound ErrorKind = "course_not_found"
	ErrorKindBlocked        ErrorKind = "blocked"
	ErrorKindChromeAccess   ErrorKind = "chrome_access"
	ErrorKindInterrupted    ErrorKind = "interrupted"
	ErrorKindOther          ErrorKind = "other"
)

var (
	ErrInvalidRequest = errors.New("invalid download request")
	ErrJobNotFound    = errors.New("job not found")
	ErrJobFinished    = errors.New("job already finished")
)

var (
	udemyHostSuffix = ".udemy.com"
	coursePath      = regexp.MustCompile(`^/course/[A-Za-z0-9_-]+/?`)
	chapterFilter   = regexp.MustCompile(`^\d+(-\d+)?(,\d+(-\d+)?)*$`)
)

type Request struct {
	CourseURL string
	Chapters  string
}

func (request Request) Host() string {
	parsed, err := url.Parse(request.CourseURL)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func (request Request) validate() error {
	parsed, err := url.Parse(request.CourseURL)
	if err != nil || parsed.Scheme != "https" || !strings.HasSuffix(parsed.Hostname(), udemyHostSuffix) || !coursePath.MatchString(parsed.Path) {
		return fmt.Errorf("%w: %q is not a https://<portal>.udemy.com/course/<name>/ URL", ErrInvalidRequest, request.CourseURL)
	}
	if request.Chapters != "" && !chapterFilter.MatchString(request.Chapters) {
		return fmt.Errorf("%w: chapters %q must look like 1,3-5", ErrInvalidRequest, request.Chapters)
	}
	return nil
}

type Progress struct {
	CourseTitle    string
	TotalLectures  int
	FinishedCount  int
	CurrentLecture string
}

type Job struct {
	ID         string
	Request    Request
	Status     Status
	CreatedAt  time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time
	Progress   Progress
	Failures   []string
	Error      string
	ErrorKind  ErrorKind
}

type EventKind string

const (
	EventCourseStarted  EventKind = "course_started"
	EventLectureStarted EventKind = "lecture_started"
	EventLectureDone    EventKind = "lecture_done"
	EventLectureFailed  EventKind = "lecture_failed"
	EventSummary        EventKind = "summary"
)

type Event struct {
	Kind          EventKind
	CourseTitle   string
	TotalLectures int
	LectureTitle  string
	Failures      []string
}

func (job *Job) apply(event Event) {
	switch event.Kind {
	case EventCourseStarted:
		job.Progress.CourseTitle, job.Progress.TotalLectures = event.CourseTitle, event.TotalLectures
	case EventLectureStarted:
		job.Progress.CurrentLecture = event.LectureTitle
	case EventLectureDone:
		job.Progress.FinishedCount++
		job.Progress.CurrentLecture = ""
	case EventLectureFailed:
		job.Progress.FinishedCount++
		job.Progress.CurrentLecture = ""
		job.Failures = append(job.Failures, event.Failures...)
	case EventSummary:
		job.Progress.CurrentLecture = ""
	}
}

func (job *Job) finish(status Status, errMessage string, kind ErrorKind, at *time.Time) {
	job.Status, job.Error, job.ErrorKind, job.FinishedAt = status, errMessage, kind, at
	job.Progress.CurrentLecture = ""
}
