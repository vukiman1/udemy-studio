package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strconv"

	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const (
	// ManifestFile is hidden, so the scanner never lists it as an attachment.
	ManifestFile   = ".course.json"
	ManifestFormat = 1
	udemyKeyPrefix = "udemy-lecture:"
)

// Manifest is what the Go downloader records about a course so a lecture keeps its identity when files are renumbered.
type Manifest struct {
	Format   int               `json:"format"`
	Lectures []ManifestLecture `json:"lectures"`
}

type ManifestLecture struct {
	UdemyID    int64  `json:"udemy_id"`
	Class      string `json:"class"`
	ChapterDir string `json:"chapter_dir"`
	File       string `json:"file"`
}

// manifestLectures maps each lecture file the manifest names, as a path under the downloads root, to its entry.
func manifestLectures(downloads fs.FS, courseDir string) (map[string]ManifestLecture, error) {
	raw, err := fs.ReadFile(downloads, path.Join(courseDir, ManifestFile))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read manifest of %q: %w", courseDir, err)
	}
	var manifest Manifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, fmt.Errorf("decode manifest of %q (delete it and download the course again to rewrite it): %w", courseDir, err)
	}
	entries := make(map[string]ManifestLecture, len(manifest.Lectures))
	for _, lecture := range manifest.Lectures {
		entries[path.Join(courseDir, lecture.ChapterDir, lecture.File)] = lecture
	}
	return entries, nil
}

func manifestEntry(lecture Lecture, entries map[string]ManifestLecture) (ManifestLecture, bool) {
	for _, file := range []string{lecture.ArticlePath, lecture.VideoPath} {
		if entry, ok := entries[file]; ok && file != "" {
			return entry, true
		}
	}
	return ManifestLecture{}, false
}

// lectureKeys returns the Udemy id key when the manifest knows the lecture's file, with the path key progress may still use.
func lectureKeys(lecture Lecture, pathKey string, entries map[string]ManifestLecture) (string, string) {
	entry, ok := manifestEntry(lecture, entries)
	if !ok {
		return pathKey, ""
	}
	return udemyKeyPrefix + strconv.FormatInt(entry.UdemyID, 10), pathKey
}

func lectureKind(lecture Lecture, entries map[string]ManifestLecture) LectureKind {
	if lecture.ArticlePath != "" && entries[lecture.ArticlePath].Class == udemy.QuizClass {
		return LectureKindQuiz
	}
	return lecture.Kind
}
