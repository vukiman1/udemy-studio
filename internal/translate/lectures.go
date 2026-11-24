package translate

import (
	"bytes"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/subtitle"
)

// PendingLectures lists a course's lectures with English subtitles and no Vietnamese ones, and counts those already translated.
func PendingLectures(course library.Course) ([]Candidate, int) {
	var pending []Candidate
	alreadyTranslated := 0
	for _, chapter := range course.Chapters {
		for _, lecture := range chapter.Lectures {
			source, hasSource := subtitleIn(lecture, SourceLanguage)
			if !hasSource {
				continue
			}
			if _, hasTarget := subtitleIn(lecture, TargetLanguage); hasTarget {
				alreadyTranslated++
				continue
			}
			pending = append(pending, Candidate{
				LectureID:    lecture.ID,
				LectureTitle: lecture.Title,
				SourcePath:   source,
				TargetPath:   strings.TrimSuffix(source, "_"+SourceLanguage+".srt") + "_" + TargetLanguage + ".srt",
			})
		}
	}
	return pending, alreadyTranslated
}

// WriteSubtitle writes beside the original under a hidden name the library scan ignores, and moves it into place only once complete.
func WriteSubtitle(files Files, targetPath string, cues []subtitle.Cue) error {
	var document bytes.Buffer
	if err := subtitle.WriteSRT(&document, cues); err != nil {
		return err
	}
	temporary := path.Join(path.Dir(targetPath), temporaryPrefix+path.Base(targetPath)+temporarySuffix)
	if err := files.WriteFile(temporary, document.Bytes(), subtitleFileMode); err != nil {
		return fmt.Errorf("write %s: %w", temporary, err)
	}
	if err := files.Rename(temporary, targetPath); err != nil {
		_ = files.Remove(temporary)
		return fmt.Errorf("move the translation into %s: %w", targetPath, err)
	}
	return nil
}

func subtitleIn(lecture library.Lecture, language string) (string, bool) {
	index := slices.IndexFunc(lecture.Subtitles, func(file library.SubtitleFile) bool { return file.Language == language })
	if index < 0 {
		return "", false
	}
	return lecture.Subtitles[index].Path, true
}
