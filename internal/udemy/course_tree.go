package udemy

import (
	"fmt"
	"slices"

	"github.com/vukiman1/udemy-studio/internal/naming"
)

const chapterClass = "chapter"

var lectureClasses = []string{"lecture", "quiz", "practice", "role-play"}

type Chapter struct {
	Index    int
	ID       int64
	Dir      string
	Lectures []Lecture
}

type Lecture struct {
	Number int
	Title  string
	Item   CurriculumItem
}

// BuildCourseTree groups the flat curriculum into chapters, numbering lectures as the Python downloader does.
func BuildCourseTree(items []CurriculumItem, continuousNumbers bool) []Chapter {
	var chapters []Chapter
	lectureNumber := 0
	for _, item := range items {
		switch {
		case item.Class == chapterClass:
			if !continuousNumbers {
				lectureNumber = 0
			}
			chapters = append(chapters, newChapter(item))
		case slices.Contains(lectureClasses, item.Class):
			lectureNumber++
			if len(chapters) == 0 {
				chapters = append(chapters, newChapter(item))
			}
			if item.ID == 0 {
				continue
			}
			current := &chapters[len(chapters)-1]
			current.Lectures = append(current.Lectures, Lecture{Number: lectureNumber, Title: lectureTitle(lectureNumber, item.Title), Item: item})
		}
	}
	return chapters
}

func newChapter(item CurriculumItem) Chapter {
	return Chapter{Index: item.ObjectIndex, ID: item.ID, Dir: fmt.Sprintf("%02d - %s", item.ObjectIndex, naming.SanitizeFilename(item.Title))}
}

func lectureTitle(number int, title string) string {
	return fmt.Sprintf("%03d %s", number, naming.SanitizeFilename(title))
}
