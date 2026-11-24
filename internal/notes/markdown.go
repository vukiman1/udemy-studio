package notes

import (
	"fmt"
	"io"
	"strings"

	"github.com/vukiman1/udemy-studio/internal/library"
)

const (
	secondsPerMinute = 60
	secondsPerHour   = 60 * secondsPerMinute
	// Continuation lines are indented to stay inside their list item.
	listContinuation = "\n  "
)

// WriteMarkdown writes a course's notes in course order, each under its chapter and lecture.
func WriteMarkdown(writer io.Writer, course library.Course, byLecture map[string][]Note) error {
	var document strings.Builder
	fmt.Fprintf(&document, "# %s\n", course.Title)
	for _, chapter := range course.Chapters {
		isChapterWritten := false
		for _, lecture := range chapter.Lectures {
			lectureNotes := byLecture[lecture.Key]
			if len(lectureNotes) == 0 {
				continue
			}
			if !isChapterWritten {
				fmt.Fprintf(&document, "\n## %02d %s\n", chapter.Index, chapter.Title)
				isChapterWritten = true
			}
			fmt.Fprintf(&document, "\n### %03d %s\n\n", lecture.Number, lecture.Title)
			for _, note := range lectureNotes {
				document.WriteString(noteItem(note, lecture.VideoPath != ""))
			}
		}
	}
	if _, err := io.WriteString(writer, document.String()); err != nil {
		return fmt.Errorf("write notes of course %q: %w", course.Title, err)
	}
	return nil
}

func noteItem(note Note, hasVideo bool) string {
	body := strings.ReplaceAll(note.Body, "\n", listContinuation)
	if !hasVideo {
		return "- " + body + "\n"
	}
	return fmt.Sprintf("- **%s** %s\n", clock(note.AtSeconds), body)
}

func clock(seconds float64) string {
	whole := int(seconds)
	hours, minutes, rest := whole/secondsPerHour, whole%secondsPerHour/secondsPerMinute, whole%secondsPerMinute
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, rest)
	}
	return fmt.Sprintf("%02d:%02d", minutes, rest)
}
