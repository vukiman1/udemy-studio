package library

import "slices"

// SubtitlesIn returns the lecture's subtitles in the order the languages are asked for, each language once; nil asks for all.
func (lecture Lecture) SubtitlesIn(languages []string) []SubtitleFile {
	if languages == nil {
		return lecture.Subtitles
	}
	var selected []SubtitleFile
	for position, language := range languages {
		if slices.Contains(languages[:position], language) {
			continue
		}
		selected = append(selected, lecture.subtitlesOf(language)...)
	}
	return selected
}

func (lecture Lecture) subtitlesOf(language string) []SubtitleFile {
	var files []SubtitleFile
	for _, file := range lecture.Subtitles {
		if file.Language == language {
			files = append(files, file)
		}
	}
	return files
}
