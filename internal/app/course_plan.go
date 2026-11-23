package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/vukiman1/udemy-studio/internal/adapter/udemyapi"
	"github.com/vukiman1/udemy-studio/internal/udemy"
)

type PlanCourseOptions struct {
	CourseURL     string
	DataDir       string
	ChromeDataDir string
	Plan          udemy.PlanOptions
}

// PlanCourse asks Udemy for a course's curriculum with the selected Chrome profile's login and plans its files.
func PlanCourse(ctx context.Context, options PlanCourseOptions, logger *slog.Logger) (udemy.Plan, error) {
	ref, err := udemy.ParseCourseURL(options.CourseURL)
	if err != nil {
		return udemy.Plan{}, err
	}
	database, err := openDatabase(ctx, options.DataDir)
	if err != nil {
		return udemy.Plan{}, err
	}
	defer database.Close()

	token, err := newLoginService(options.ChromeDataDir, database).Token(ctx, ref.Host())
	if err != nil {
		return udemy.Plan{}, err
	}
	client, err := udemyapi.New(ref.Portal, token, logger)
	if err != nil {
		return udemy.Plan{}, err
	}
	if err := client.Visit(ctx); err != nil {
		return udemy.Plan{}, err
	}
	course, err := client.ResolveCourse(ctx, options.CourseURL, ref)
	if err != nil {
		return udemy.Plan{}, err
	}
	logger.Info("course found", "id", course.ID, "title", course.Title)
	items, err := client.Curriculum(ctx, course.ID)
	if err != nil {
		return udemy.Plan{}, err
	}
	plan, err := udemy.BuildPlan(ctx, course, items, options.Plan, client)
	if err != nil {
		return udemy.Plan{}, fmt.Errorf("plan course %d: %w", course.ID, err)
	}
	return plan, nil
}
