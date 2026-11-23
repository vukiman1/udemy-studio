package app

import (
	"context"
	"log/slog"

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
	client, err := courseClient(ctx, options.DataDir, options.ChromeDataDir, ref, logger)
	if err != nil {
		return udemy.Plan{}, err
	}
	return planCourse(ctx, client, ref, options.Plan, logger)
}

func planCourse(ctx context.Context, client udemy.CourseSource, ref udemy.CourseRef, options udemy.PlanOptions, logger *slog.Logger) (udemy.Plan, error) {
	plan, err := udemy.PlanCourse(ctx, client, ref, options)
	if err != nil {
		return udemy.Plan{}, err
	}
	logger.Info("course planned", "course", plan.CourseName, "chapters", len(plan.Chapters), "warnings", len(plan.Warnings))
	return plan, nil
}
