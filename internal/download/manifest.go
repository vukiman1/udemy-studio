package download

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/udemy"
)

// writeManifest records every planned lecture's Udemy id.
func writeManifest(courseDir string, plan udemy.Plan) error {
	manifest := library.Manifest{Format: library.ManifestFormat, Lectures: []library.ManifestLecture{}}
	for _, chapter := range plan.Chapters {
		for _, lecture := range chapter.Lectures {
			manifest.Lectures = append(manifest.Lectures, library.ManifestLecture{UdemyID: lecture.ID, Class: lecture.Class, ChapterDir: chapter.Dir, File: lecture.File})
		}
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode course manifest: %w", err)
	}
	if err := writeLibraryFile(filepath.Join(courseDir, library.ManifestFile), encoded); err != nil {
		return fmt.Errorf("save course manifest: %w", err)
	}
	return nil
}
