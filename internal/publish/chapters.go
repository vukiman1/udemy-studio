package publish

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var (
	ErrNoChapters     = errors.New("choose at least one chapter to publish")
	ErrUnknownChapter = errors.New("no such chapter in the course")
)

type ChapterPlan struct {
	Dir         string
	Files       int
	SizeBytes   int64
	IsPublished bool
}

type Selection struct {
	Files      []courseFile
	TotalBytes int64
	Chapters   []string
}

func chapterOf(filePath string) string {
	chapter, _, isInFolder := strings.Cut(filePath, "/")
	if !isInFolder {
		return ""
	}
	return chapter
}

func (plan Plan) Chapters() []ChapterPlan {
	var chapters []ChapterPlan
	for _, file := range plan.Files {
		dir := chapterOf(file.Path)
		if dir == "" {
			continue
		}
		if len(chapters) == 0 || chapters[len(chapters)-1].Dir != dir {
			chapters = append(chapters, ChapterPlan{Dir: dir, IsPublished: plan.isChapterPublished(dir)})
		}
		chapters[len(chapters)-1].Files++
		chapters[len(chapters)-1].SizeBytes += file.SizeBytes
	}
	return chapters
}

func (plan Plan) SharedBytes() int64 {
	var shared int64
	for _, file := range plan.Files {
		if chapterOf(file.Path) == "" {
			shared += file.SizeBytes
		}
	}
	return shared
}

func (plan Plan) Select(chapters []string) (Selection, error) {
	if len(chapters) == 0 {
		return Selection{}, ErrNoChapters
	}
	known := plan.Chapters()
	for _, chapter := range chapters {
		if !slices.ContainsFunc(known, func(candidate ChapterPlan) bool { return candidate.Dir == chapter }) {
			return Selection{}, fmt.Errorf("%w: %q", ErrUnknownChapter, chapter)
		}
	}
	selection := Selection{}
	for _, candidate := range known {
		if slices.Contains(chapters, candidate.Dir) {
			selection.Chapters = append(selection.Chapters, candidate.Dir)
		}
	}
	for _, file := range plan.Files {
		chapter := chapterOf(file.Path)
		if chapter == "" || slices.Contains(chapters, chapter) {
			selection.Files = append(selection.Files, file)
			selection.TotalBytes += file.SizeBytes
		}
	}
	return selection, nil
}

func (plan Plan) isChapterPublished(dir string) bool {
	if plan.Published == nil || plan.Published.Course.Dir != plan.Course.Dir {
		return false
	}
	return len(plan.Published.Chapters) == 0 || slices.Contains(plan.Published.Chapters, dir)
}
