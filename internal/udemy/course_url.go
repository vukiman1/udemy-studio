package udemy

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const (
	udemyHostSuffix = ".udemy.com"
	coursePathPart  = "course"
	draftPathPart   = "draft"
)

var (
	ErrInvalidCourseURL = errors.New("not a Udemy course URL")
	courseSlugPattern   = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

type CourseRef struct {
	Portal  string
	Slug    string
	IsDraft bool
}

func (ref CourseRef) Host() string {
	return ref.Portal + udemyHostSuffix
}

// ParseCourseURL splits https://<portal>.udemy.com/course/<slug-or-id>/... into its portal and slug.
func ParseCourseURL(raw string) (CourseRef, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return CourseRef{}, fmt.Errorf("%w: %w", ErrInvalidCourseURL, err)
	}
	portal, isUdemy := strings.CutSuffix(strings.ToLower(parsed.Hostname()), udemyHostSuffix)
	if !isUdemy || portal == "" || strings.Contains(portal, ".") {
		return CourseRef{}, fmt.Errorf("%w: host %q is not <portal>%s", ErrInvalidCourseURL, parsed.Hostname(), udemyHostSuffix)
	}
	parts := strings.FieldsFunc(parsed.Path, func(r rune) bool { return r == '/' })
	isDraft := false
	if len(parts) > 0 && parts[0] == coursePathPart {
		parts = parts[1:]
		for len(parts) > 0 && parts[0] == draftPathPart {
			parts, isDraft = parts[1:], true
		}
	}
	if len(parts) == 0 || !courseSlugPattern.MatchString(parts[0]) {
		return CourseRef{}, fmt.Errorf("%w: no course slug in %q", ErrInvalidCourseURL, parsed.Path)
	}
	return CourseRef{Portal: portal, Slug: parts[0], IsDraft: isDraft}, nil
}
