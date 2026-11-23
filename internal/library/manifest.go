package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strconv"
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
	ChapterDir string `json:"chapter_dir"`
	File       string `json:"file"`
}

// lectureIDs maps each lecture file the manifest names, as a path under the downloads root, to its Udemy id.
func lectureIDs(downloads fs.FS, courseDir string) (map[string]int64, error) {
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
	ids := make(map[string]int64, len(manifest.Lectures))
	for _, lecture := range manifest.Lectures {
		ids[path.Join(courseDir, lecture.ChapterDir, lecture.File)] = lecture.UdemyID
	}
	return ids, nil
}

// lectureKey is the key progress is saved under: the Udemy id when the manifest knows the lecture's file, else its path.
func lectureKey(lecture Lecture, pathKey string, ids map[string]int64) string {
	for _, file := range []string{lecture.ArticlePath, lecture.VideoPath} {
		if id, ok := ids[file]; ok && file != "" {
			return udemyKeyPrefix + strconv.FormatInt(id, 10)
		}
	}
	return pathKey
}
