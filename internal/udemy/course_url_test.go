package udemy

import (
	"errors"
	"testing"
)

func TestParseCourseURL(t *testing.T) {
	invalid := CourseRef{}
	cases := []struct {
		raw  string
		want CourseRef
	}{
		{"https://example-portal.udemy.com/course/backend-basics/learn/lecture/1#overview", CourseRef{Portal: "example-portal", Slug: "backend-basics"}},
		{"https://www.udemy.com/course/draft/123456/", CourseRef{Portal: "www", Slug: "123456", IsDraft: true}},
		{"https://WWW.Udemy.com/backend-basics", CourseRef{Portal: "www", Slug: "backend-basics"}},
		{"https://udemy.com/course/x", invalid},
		{"https://evil.example/course/x", invalid},
		{"https://a.b.udemy.com/course/x", invalid},
		{"https://www.udemy.com/course/", invalid},
	}
	for _, testCase := range cases {
		got, err := ParseCourseURL(testCase.raw)
		if got != testCase.want || (testCase.want == invalid) != errors.Is(err, ErrInvalidCourseURL) {
			t.Errorf("ParseCourseURL(%q) = %+v, %v; want %+v", testCase.raw, got, err, testCase.want)
		}
	}
}
