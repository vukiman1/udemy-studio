package publish

import (
	"fmt"
	"io/fs"
	"mime"
	"path"
	"strings"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
)

const fallbackContentType = "application/octet-stream"

var (
	unfinishedSuffixes = []string{".aria2", ".part", ".partial", ".tmp", ".ytdl"}
	contentTypes       = map[string]string{
		".mp4":  "video/mp4",
		".vtt":  "text/vtt; charset=utf-8",
		".srt":  "application/x-subrip",
		".html": "text/html; charset=utf-8",
		".json": "application/json",
		".jpg":  "image/jpeg",
		".pdf":  "application/pdf",
	}
)

type courseFile struct {
	Path       string
	SizeBytes  int64
	ModifiedAt time.Time
}

func listCourseFiles(files fs.FS, courseDir string) ([]courseFile, error) {
	var listed []courseFile
	err := fs.WalkDir(files, courseDir, func(walked string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return skipHiddenDir(walked, courseDir, entry.Name())
		}
		if !isPublishable(entry.Name()) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		listed = append(listed, courseFile{Path: strings.TrimPrefix(walked, courseDir+"/"), SizeBytes: info.Size(), ModifiedAt: info.ModTime()})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list course %s: %w", courseDir, err)
	}
	return listed, nil
}

func skipHiddenDir(walked, courseDir, name string) error {
	if walked != courseDir && strings.HasPrefix(name, ".") {
		return fs.SkipDir
	}
	return nil
}

func isPublishable(name string) bool {
	if name == library.ManifestFile {
		return true
	}
	if strings.HasPrefix(name, ".") {
		return false
	}
	for _, suffix := range unfinishedSuffixes {
		if strings.HasSuffix(name, suffix) {
			return false
		}
	}
	return true
}

func contentType(name string) string {
	extension := strings.ToLower(path.Ext(name))
	if known, found := contentTypes[extension]; found {
		return known
	}
	if guessed := mime.TypeByExtension(extension); guessed != "" {
		return guessed
	}
	return fallbackContentType
}
