package udemy

import (
	"fmt"
	"strconv"

	"github.com/vukiman1/udemy-studio/internal/naming"
)

const (
	videoExtension = ".mp4"
	htmlExtension  = ".html"
	srtExtension   = ".srt"
	vttExtension   = ".vtt"
	urlExtension   = ".url"
)

func courseDirName(course Course, idAsCourseName bool) string {
	if idAsCourseName {
		return naming.SanitizeFilename(strconv.FormatInt(course.ID, 10))
	}
	return naming.SanitizeFilename(course.Slug)
}

func videoFileName(lectureTitle string) string {
	return naming.RemoveEmoji(naming.SanitizeFilename(lectureTitle + videoExtension))
}

func htmlFileName(lectureTitle string) string {
	return naming.SanitizeFilename(lectureTitle) + htmlExtension
}

func quizFileName(lectureTitle string) string {
	return naming.SanitizeFilename(lectureTitle + htmlExtension)
}

func captionStem(lectureTitle, language string) string {
	return naming.SanitizeFilename(lectureTitle) + "_" + naming.SanitizeFilename(language)
}

func assetFileName(lectureNumber int, filename string) string {
	return naming.SanitizeFilename(fmt.Sprintf("%03d %s", lectureNumber, filename))
}
