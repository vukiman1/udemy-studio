package udemy

import (
	"errors"
	"testing"
)

func TestParseCourseURL(t *testing.T) {
	cases := []struct {
		raw  string
		want CourseRef
	}{
		{"https://example-portal.udemy.com/course/backend-basics/learn/lecture/1#overview", CourseRef{"example-portal", "backend-basics"}},
		{"https://www.udemy.com/course/draft/123456/", CourseRef{"www", "123456"}},
		{"https://WWW.Udemy.com/backend-basics", CourseRef{"www", "backend-basics"}},
	}
	for _, testCase := range cases {
		got, err := ParseCourseURL(testCase.raw)
		if err != nil || got != testCase.want {
			t.Errorf("ParseCourseURL(%q) = %+v, %v; want %+v", testCase.raw, got, err, testCase.want)
		}
	}
}

func TestParseCourseURLRejectsOtherHosts(t *testing.T) {
	for _, raw := range []string{"https://udemy.com/course/x", "https://evil.example/course/x", "https://a.b.udemy.com/course/x", "https://www.udemy.com/course/"} {
		if _, err := ParseCourseURL(raw); !errors.Is(err, ErrInvalidCourseURL) {
			t.Errorf("ParseCourseURL(%q) error = %v, want ErrInvalidCourseURL", raw, err)
		}
	}
}
