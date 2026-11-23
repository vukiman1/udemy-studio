package download

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// Library files are readable like the ones aria2c and yt-dlp write, as in the Python downloader.
	libraryFileMode = 0o644
	// Hidden, so the library never lists a write that was cut short.
	partialFilePattern = ".udemy-studio-*.partial"
)

// writeLibraryFile replaces path only once data is fully written, so an interrupted run never leaves a file that looks done.
func writeLibraryFile(path string, data []byte) error {
	partial, err := os.CreateTemp(filepath.Dir(path), partialFilePattern)
	if err != nil {
		return fmt.Errorf("create a temporary file for %s: %w", filepath.Base(path), err)
	}
	_, writeErr := partial.Write(data)
	if err := errors.Join(writeErr, partial.Chmod(libraryFileMode), partial.Close()); err != nil {
		return errors.Join(fmt.Errorf("write %s: %w", filepath.Base(path), err), os.Remove(partial.Name()))
	}
	if err := os.Rename(partial.Name(), path); err != nil {
		return errors.Join(fmt.Errorf("save %s: %w", filepath.Base(path), err), os.Remove(partial.Name()))
	}
	return nil
}
