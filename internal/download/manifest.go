package download

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const manifestTempSuffix = ".tmp"

// writeManifest records every planned lecture's Udemy id, replacing the old manifest only once the new one is complete.
func writeManifest(courseDir string, plan udemy.Plan) error {
	manifest := library.Manifest{Format: library.ManifestFormat, Lectures: []library.ManifestLecture{}}
	for _, chapter := range plan.Chapters {
		for _, lecture := range chapter.Lectures {
			manifest.Lectures = append(manifest.Lectures, library.ManifestLecture{UdemyID: lecture.ID, ChapterDir: chapter.Dir, File: lecture.File})
		}
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode course manifest: %w", err)
	}
	manifestPath := filepath.Join(courseDir, library.ManifestFile)
	if err := writeLibraryFile(manifestPath+manifestTempSuffix, encoded); err != nil {
		return fmt.Errorf("write course manifest: %w", err)
	}
	if err := os.Rename(manifestPath+manifestTempSuffix, manifestPath); err != nil {
		return fmt.Errorf("save course manifest: %w", err)
	}
	return nil
}
