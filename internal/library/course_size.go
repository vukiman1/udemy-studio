package library

import (
	"fmt"
	"io/fs"
)

// SizeBytes adds up every file in the course's folder: videos, subtitles, attachments and anything else downloaded
// with it. It reads the folder each time, so the size follows a compression or a download without a rescan.
func SizeBytes(downloads fs.FS, course Course) (int64, error) {
	var total int64
	err := fs.WalkDir(downloads, course.Dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("measure course %s: %w", course.Dir, err)
	}
	return total, nil
}
