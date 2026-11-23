package udemyapi

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const (
	courseSearchPath   = "/api-2.0/users/me/subscribed-courses?fields[course]=id,url,title,published_title&page=1&page_size=500&search="
	myCoursesPath      = "/api-2.0/users/me/subscribed-courses?fields[course]=id,url,title,published_title&ordering=-last_accessed,-access_time&page=1&page_size=10000"
	archivedSuffix     = "&is_archived=true"
	subscriptionPath   = "/api-2.0/users/me/subscription-course-enrollments?fields%5Buser%5D=title%2Cimage_100x100&fields%5Bcourse%5D=title%2Cheadline%2Curl%2Ccompletion_ratio%2Cnum_published_lectures%2Cimage_480x270%2Cimage_240x135%2Cfavorite_time%2Carchive_time%2Cis_taking_disabled%2Cfeatures%2Cvisible_instructors%2Clast_accessed_time%2Csort_order%2Cis_user_subscribed%2Cis_in_user_subscription%2Cis_wishlisted%2Cpublished_title%2Cavailable_features%2Cnum_published_practice_tests%2Cnum_coding_exercises%2Cnum_published_quizzes%2Cnum_of_published_curriculum_objects%2Cprimary_category%2Clocale%2Ccourse_has_labels%2Cis_gen_ai_policy_opted_in%2Cavailable_features&ordering=-last_accessed%2C-enrolled&page=1&page_size=50&locale=en_US"
	collectionsPath    = "/api-2.0/users/me/subscribed-courses-collections/?collection_has_courses=True&course_limit=20&fields[course]=last_accessed_time,title,published_title&fields[user_has_subscribed_courses_collection]=@all&page=1&page_size=1000"
	coursePathTemplate = "/api-2.0/courses/%d/"
)

var ErrCourseNotFound = errors.New("course not found: you may not be enrolled, or need to open the course once in the browser while logged in")

type courseLookup struct {
	name string
	list func(ctx context.Context) ([]udemy.Course, error)
}

// ResolveCourse finds the enrolled course a URL points to, trying each place Udemy lists courses before the course page itself.
func (client *Client) ResolveCourse(ctx context.Context, ref udemy.CourseRef) (udemy.Course, error) {
	for _, lookup := range client.courseLookups(ref.Slug) {
		courses, err := lookup.list(ctx)
		var statusErr *StatusError
		if errors.As(err, &statusErr) {
			client.logger.Warn("course lookup failed, trying the next one", "lookup", lookup.name, "err", err)
			continue
		}
		if err != nil {
			return udemy.Course{}, fmt.Errorf("look up course %q in %s: %w", ref.Slug, lookup.name, err)
		}
		if course, ok := matchCourse(courses, ref.Slug); ok {
			course.Portal = ref.Portal
			return course, nil
		}
	}
	course, err := client.courseFromPage(ctx, ref)
	var statusErr *StatusError
	if errors.Is(err, errNoCourseData) || errors.As(err, &statusErr) {
		return udemy.Course{}, fmt.Errorf("%w (%q): %w", ErrCourseNotFound, ref.Slug, err)
	}
	if err != nil {
		return udemy.Course{}, fmt.Errorf("look up course %q from its page: %w", ref.Slug, err)
	}
	course.Portal = ref.Portal
	return course, nil
}

func (client *Client) courseLookups(slug string) []courseLookup {
	return []courseLookup{
		{"subscribed course search", func(ctx context.Context) ([]udemy.Course, error) {
			return firstPage[udemy.Course](ctx, client, client.portalURL+courseSearchPath+url.QueryEscape(slug))
		}},
		{"my courses", func(ctx context.Context) ([]udemy.Course, error) {
			return allPages[udemy.Course](ctx, client, client.portalURL+myCoursesPath)
		}},
		{"subscription courses", func(ctx context.Context) ([]udemy.Course, error) {
			return allPages[udemy.Course](ctx, client, client.portalURL+subscriptionPath)
		}},
		{"course collections", client.collectionCourses},
		{"archived courses", func(ctx context.Context) ([]udemy.Course, error) {
			return firstPage[udemy.Course](ctx, client, client.portalURL+myCoursesPath+archivedSuffix)
		}},
	}
}

func (client *Client) collectionCourses(ctx context.Context) ([]udemy.Course, error) {
	type collection struct {
		Courses []udemy.Course `json:"courses"`
	}
	collections, err := firstPage[collection](ctx, client, client.portalURL+collectionsPath)
	if err != nil {
		return nil, err
	}
	var courses []udemy.Course
	for _, entry := range collections {
		courses = append(courses, entry.Courses...)
	}
	return courses, nil
}

func (client *Client) courseByID(ctx context.Context, courseID int64) (udemy.Course, error) {
	var course udemy.Course
	if err := client.getJSON(ctx, client.portalURL+fmt.Sprintf(coursePathTemplate, courseID), &course); err != nil {
		return udemy.Course{}, fmt.Errorf("fetch course %d: %w", courseID, err)
	}
	return course, nil
}

func matchCourse(courses []udemy.Course, slug string) (udemy.Course, bool) {
	for _, course := range courses {
		if course.Slug == slug || strconv.FormatInt(course.ID, 10) == slug {
			return course, true
		}
	}
	return udemy.Course{}, false
}
