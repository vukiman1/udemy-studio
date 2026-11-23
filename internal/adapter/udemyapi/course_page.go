package udemyapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"golang.org/x/net/html"

	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const (
	courseTakingAppClass = "ud-component--course-taking--app"
	moduleArgsAttribute  = "data-module-args"
)

var errNoCourseData = errors.New("the course page has no course data: the URL should be the course's /learn page and the profile must be logged in")

func (client *Client) courseFromPage(ctx context.Context, courseURL string) (udemy.Course, error) {
	page, err := client.getText(ctx, courseURL)
	if err != nil {
		return udemy.Course{}, fmt.Errorf("fetch course page: %w", err)
	}
	courseID, err := courseIDFromPage(page)
	if err != nil {
		return udemy.Course{}, err
	}
	return client.courseByID(ctx, courseID)
}

func courseIDFromPage(page string) (int64, error) {
	document, err := html.Parse(strings.NewReader(page))
	if err != nil {
		return 0, fmt.Errorf("parse course page: %w", err)
	}
	for node := range document.Descendants() {
		if node.Type != html.ElementNode || node.Data != "div" || !hasClass(node, courseTakingAppClass) {
			continue
		}
		var moduleArgs struct {
			CourseID int64 `json:"courseId"`
		}
		if err := json.Unmarshal([]byte(attribute(node, moduleArgsAttribute)), &moduleArgs); err != nil {
			return 0, fmt.Errorf("decode %s of the course page: %w", moduleArgsAttribute, err)
		}
		if moduleArgs.CourseID == 0 {
			return 0, errNoCourseData
		}
		return moduleArgs.CourseID, nil
	}
	return 0, errNoCourseData
}

func hasClass(node *html.Node, class string) bool {
	return slices.Contains(strings.Fields(attribute(node, "class")), class)
}

func attribute(node *html.Node, name string) string {
	for _, attr := range node.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}
