package udemy

import (
	"context"
	"fmt"
)

type CourseSource interface {
	ResolveCourse(ctx context.Context, ref CourseRef) (Course, error)
	Curriculum(ctx context.Context, courseID int64) ([]CurriculumItem, error)
	PlaylistFetcher
}

// PlanCourse finds the course a reference points to and plans every file it becomes.
func PlanCourse(ctx context.Context, source CourseSource, ref CourseRef, options PlanOptions) (Plan, error) {
	course, err := source.ResolveCourse(ctx, ref)
	if err != nil {
		return Plan{}, err
	}
	items, err := source.Curriculum(ctx, course.ID)
	if err != nil {
		return Plan{}, err
	}
	plan, err := BuildPlan(ctx, course, items, options, source)
	if err != nil {
		return Plan{}, fmt.Errorf("plan course %d: %w", course.ID, err)
	}
	return plan, nil
}
