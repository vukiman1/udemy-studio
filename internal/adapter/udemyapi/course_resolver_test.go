package udemyapi

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const courseURL = testPortalURL + "/course/backend-basics/learn/"

var courseRef = udemy.CourseRef{Portal: "example-portal", Slug: "backend-basics"}

func TestCourseFoundBySearch(t *testing.T) {
	fake := &fakeTransport{}
	fake.reply(testPortalURL+courseSearchPath+"backend-basics", http.StatusOK, `{"results": [{"id": 9, "published_title": "other"}, {"id": 42, "title": "Backend", "published_title": "backend-basics"}]}`)
	course, err := newTestClient(t, fake).ResolveCourse(t.Context(), courseURL, courseRef)
	want := udemy.Course{ID: 42, Title: "Backend", Slug: "backend-basics", Portal: "example-portal"}
	if err != nil || course != want {
		t.Errorf("ResolveCourse = %+v, %v; want %+v", course, err, want)
	}
}

func TestCourseFoundInCollectionsAfterEarlierLookupsMiss(t *testing.T) {
	fake := &fakeTransport{}
	fake.reply(testPortalURL+courseSearchPath+"backend-basics", http.StatusOK, `{"results": []}`)
	fake.reply(testPortalURL+myCoursesPath, http.StatusOK, `{"results": [{"id": 1, "published_title": "x"}]}`)
	fake.reply(testPortalURL+subscriptionPath, http.StatusForbidden, `{"detail": "Not a subscriber."}`)
	fake.reply(testPortalURL+collectionsPath, http.StatusOK, `{"results": [{"courses": [{"id": 42, "published_title": "backend-basics"}]}]}`)
	course, err := newTestClient(t, fake).ResolveCourse(t.Context(), courseURL, courseRef)
	if err != nil || course.ID != 42 || fake.requestedMatching("is_archived") != 0 {
		t.Errorf("ResolveCourse = %+v, %v", course, err)
	}
}

func TestCourseMatchedByNumericID(t *testing.T) {
	fake := &fakeTransport{}
	fake.reply(testPortalURL+courseSearchPath+"42", http.StatusOK, `{"results": [{"id": 42, "published_title": "backend-basics"}]}`)
	course, err := newTestClient(t, fake).ResolveCourse(t.Context(), courseURL, udemy.CourseRef{Portal: "example-portal", Slug: "42"})
	if err != nil || course.ID != 42 {
		t.Errorf("ResolveCourse = %+v, %v", course, err)
	}
}

func TestCourseFoundFromCoursePageAsLastResort(t *testing.T) {
	fake := &fakeTransport{}
	fake.reply(courseURL, http.StatusOK, `<html><body><div class="main ud-component--course-taking--app" data-module-args="{&quot;courseId&quot;: 42}"></div></body></html>`)
	fake.reply(testPortalURL+"/api-2.0/courses/42/", http.StatusOK, `{"id": 42, "title": "Backend", "published_title": "backend-basics"}`)
	course, err := newTestClient(t, fake).ResolveCourse(t.Context(), courseURL, courseRef)
	if err != nil || course.ID != 42 || course.Portal != "example-portal" {
		t.Errorf("ResolveCourse = %+v, %v", course, err)
	}
}

func TestUnknownCourseIsNotFound(t *testing.T) {
	fake := &fakeTransport{}
	fake.reply(courseURL, http.StatusOK, `<html><body>Log in</body></html>`)
	if _, err := newTestClient(t, fake).ResolveCourse(t.Context(), courseURL, courseRef); !errors.Is(err, ErrCourseNotFound) {
		t.Errorf("err = %v, want ErrCourseNotFound", err)
	}
}

func TestNetworkErrorsAbortTheLookup(t *testing.T) {
	fake := &fakeTransport{}
	fake.reply(testPortalURL+courseSearchPath+"backend-basics", http.StatusOK, `<html>blocked</html>`)
	var unexpected *UnexpectedResponseError
	if _, err := newTestClient(t, fake).ResolveCourse(t.Context(), courseURL, courseRef); !errors.As(err, &unexpected) || fake.requestedMatching("page_size=10000") != 0 {
		t.Errorf("err = %v", err)
	}
}

func TestCoursePageErrorStatusIsNotFound(t *testing.T) {
	if _, err := newTestClient(t, &fakeTransport{}).ResolveCourse(t.Context(), courseURL, courseRef); !errors.Is(err, ErrCourseNotFound) {
		t.Errorf("err = %v, want ErrCourseNotFound", err)
	}
}

func TestCancelledCoursePageFetchIsNotReportedAsNotFound(t *testing.T) {
	fake := &fakeTransport{}
	fake.reply(courseURL, http.StatusServiceUnavailable, "busy")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := newTestClient(t, fake).ResolveCourse(ctx, courseURL, courseRef)
	if !errors.Is(err, context.Canceled) || errors.Is(err, ErrCourseNotFound) {
		t.Errorf("err = %v, want context.Canceled only", err)
	}
}
