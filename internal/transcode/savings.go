package transcode

import (
	"context"
	"fmt"
	"strings"
)

type Saving struct {
	Path        string
	CourseDir   string
	BytesBefore int64
	BytesAfter  int64
}

type SizeTotals struct {
	BytesBefore int64
	BytesAfter  int64
}

type SavingStore interface {
	Save(ctx context.Context, saving Saving) error
	CourseTotals(ctx context.Context, courseDir string) (SizeTotals, error)
	AllTotals(ctx context.Context) (SizeTotals, error)
}

func (totals SizeTotals) shareAfter() float64 {
	if totals.BytesBefore == 0 {
		return 0
	}
	return float64(totals.BytesAfter) / float64(totals.BytesBefore)
}

func (service *Service) expectedShareAfter(ctx context.Context, courseDir string) (float64, error) {
	course, err := service.deps.Savings.CourseTotals(ctx, courseDir)
	if err != nil {
		return 0, fmt.Errorf("sum the savings of %s: %w", courseDir, err)
	}
	if course.BytesBefore > 0 {
		return course.shareAfter(), nil
	}
	all, err := service.deps.Savings.AllTotals(ctx)
	if err != nil {
		return 0, fmt.Errorf("sum all savings: %w", err)
	}
	return all.shareAfter(), nil
}

func (service *Service) recordSaving(ctx context.Context, videoPath string, bytesBefore, bytesAfter int64) {
	courseDir, _, _ := strings.Cut(videoPath, "/")
	saving := Saving{Path: videoPath, CourseDir: courseDir, BytesBefore: bytesBefore, BytesAfter: bytesAfter}
	if err := service.deps.Savings.Save(ctx, saving); err != nil {
		service.deps.Logger.Warn("this compression will not inform later estimates", "path", videoPath, "err", err)
	}
}
