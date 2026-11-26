package transcode

import (
	"errors"
	"fmt"
	"path"
	"slices"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
)

var (
	ErrNoChapters     = errors.New("choose at least one chapter to compress")
	ErrUnknownChapter = errors.New("no such chapter in the course")
)

type ChapterPlan struct {
	Dir          string
	Videos       int
	TotalBytes   int64
	EstimatedFor time.Duration
}

func chapterOf(videoPath string) string {
	return path.Base(path.Dir(videoPath))
}

func (plan Plan) Chapters() []ChapterPlan {
	var chapters []ChapterPlan
	var lengths []time.Duration
	for _, candidate := range plan.Candidates {
		dir := chapterOf(candidate.Path)
		if len(chapters) == 0 || chapters[len(chapters)-1].Dir != dir {
			chapters = append(chapters, ChapterPlan{Dir: dir})
			lengths = append(lengths, 0)
		}
		last := len(chapters) - 1
		chapters[last].Videos++
		chapters[last].TotalBytes += candidate.Media.SizeBytes
		lengths[last] += candidate.Media.Duration
	}
	for position := range chapters {
		chapters[position].EstimatedFor = time.Duration(float64(lengths[position]) / assumedRealtimeSpeed)
	}
	return chapters
}

func (plan Plan) only(chapters []string) []Candidate {
	return slices.DeleteFunc(slices.Clone(plan.Candidates), func(candidate Candidate) bool {
		return !slices.Contains(chapters, chapterOf(candidate.Path))
	})
}

func checkChapters(course library.Course, chapters []string) error {
	if len(chapters) == 0 {
		return ErrNoChapters
	}
	known := map[string]bool{}
	for _, chapter := range course.Chapters {
		for _, lecture := range chapter.Lectures {
			if lecture.VideoPath != "" {
				known[chapterOf(lecture.VideoPath)] = true
			}
		}
	}
	for _, chapter := range chapters {
		if !known[chapter] {
			return fmt.Errorf("%w: %q", ErrUnknownChapter, chapter)
		}
	}
	return nil
}
