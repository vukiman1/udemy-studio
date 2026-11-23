package download

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/vukiman1/udemy-studio/internal/udemy"
)

var ErrInvalidChapterFilter = errors.New("invalid chapter filter")

// ChapterFilter holds the chapter ranges to download; nil downloads every chapter.
type ChapterFilter []chapterRange

type chapterRange struct {
	first, last int
}

// ParseChapterFilter reads a list like "1,3-5,7"; an empty one selects every chapter.
func ParseChapterFilter(text string) (ChapterFilter, error) {
	if text == "" {
		return nil, nil
	}
	var filter ChapterFilter
	for part := range strings.SplitSeq(text, ",") {
		chapters, err := parseChapterRange(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		filter = append(filter, chapters)
	}
	return filter, nil
}

func parseChapterRange(part string) (chapterRange, error) {
	first, last, isRange := strings.Cut(part, "-")
	if !isRange {
		last = first
	}
	start, startErr := strconv.Atoi(strings.TrimSpace(first))
	end, endErr := strconv.Atoi(strings.TrimSpace(last))
	if startErr != nil || endErr != nil || start > end {
		return chapterRange{}, fmt.Errorf("%w: %q, expected e.g. 1,3-5,7", ErrInvalidChapterFilter, part)
	}
	return chapterRange{first: start, last: end}, nil
}

func (filter ChapterFilter) includes(index int) bool {
	return filter == nil || slices.ContainsFunc(filter, func(chapters chapterRange) bool {
		return index >= chapters.first && index <= chapters.last
	})
}

func (filter ChapterFilter) lectureCount(plan udemy.Plan) int {
	count := 0
	for _, chapter := range plan.Chapters {
		if !filter.includes(chapter.Index) {
			continue
		}
		for _, lecture := range chapter.Lectures {
			if lecture.Kind != udemy.KindQuiz {
				count++
			}
		}
	}
	return count
}
