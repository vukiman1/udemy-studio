package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/subtitle"
	"github.com/vukiman1/udemy-studio/internal/translate"
)

const (
	handoffManifestName = "handoff.json"
	sourceTextSuffix    = ".en.txt"
	translatedSuffix    = ".vi.txt"
	handoffFileMode     = 0o600
	handoffDirMode      = 0o750
)

var ErrCourseNotInLibrary = errors.New("no course in the library matches")

// SubtitleHandoff moves subtitles between the library and a work directory, where they are translated outside the app.
type SubtitleHandoff struct {
	LibraryDir string
	WorkDir    string
}

type handoffEntry struct {
	File         string `json:"file"`
	Course       string `json:"course"`
	LectureTitle string `json:"lecture"`
	SourcePath   string `json:"source"`
	TargetPath   string `json:"target"`
}

type ImportReport struct {
	Written  []string
	Rejected []string
	Waiting  []string
}

// Export writes one "<index>\t<text>" file per lecture still lacking Vietnamese subtitles, in course order in the
// manifest; it never writes into the library.
func (handoff SubtitleHandoff) Export(course string) (int, error) {
	scanned, err := library.Scan(os.DirFS(handoff.LibraryDir))
	if err != nil {
		return 0, fmt.Errorf("scan library %s: %w", handoff.LibraryDir, err)
	}
	courses, err := matchingCourses(scanned.Courses, course)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(handoff.WorkDir, handoffDirMode); err != nil {
		return 0, fmt.Errorf("create work dir %s: %w", handoff.WorkDir, err)
	}
	var entries []handoffEntry
	for _, matched := range courses {
		pending, _ := translate.PendingLectures(matched)
		for _, candidate := range pending {
			// Files are named by the lecture's stable id, so exporting again never pairs an old translation with another lecture.
			entry := handoffEntry{File: candidate.LectureID, Course: matched.Title, LectureTitle: candidate.LectureTitle, SourcePath: candidate.SourcePath, TargetPath: candidate.TargetPath}
			if err := handoff.exportLecture(entry); err != nil {
				return 0, err
			}
			entries = append(entries, entry)
		}
	}
	manifest, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("encode %s: %w", handoffManifestName, err)
	}
	if err := os.WriteFile(filepath.Join(handoff.WorkDir, handoffManifestName), manifest, handoffFileMode); err != nil {
		return 0, fmt.Errorf("write %s: %w", handoffManifestName, err)
	}
	return len(entries), nil
}

func (handoff SubtitleHandoff) exportLecture(entry handoffEntry) error {
	cues, err := subtitle.ParseSRTFile(os.DirFS(handoff.LibraryDir), entry.SourcePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", entry.SourcePath, err)
	}
	path := filepath.Join(handoff.WorkDir, entry.File+sourceTextSuffix)
	if err := os.WriteFile(path, []byte(translate.CueLines(cues)), handoffFileMode); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// Import writes every translation that passes the app's own checks into the library and reports the rest.
func (handoff SubtitleHandoff) Import() (ImportReport, error) {
	raw, err := os.ReadFile(filepath.Join(handoff.WorkDir, handoffManifestName)) //nolint:gosec // reason: the work dir is the operator's --work flag
	if err != nil {
		return ImportReport{}, fmt.Errorf("read %s from the work dir: %w", handoffManifestName, err)
	}
	var entries []handoffEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return ImportReport{}, fmt.Errorf("decode %s: %w", handoffManifestName, err)
	}
	media, err := os.OpenRoot(handoff.LibraryDir)
	if err != nil {
		return ImportReport{}, fmt.Errorf("open library %s: %w", handoff.LibraryDir, err)
	}
	defer media.Close()
	var report ImportReport
	for _, entry := range entries {
		label := entry.File + " " + entry.LectureTitle
		translation, err := os.ReadFile(filepath.Join(handoff.WorkDir, entry.File+translatedSuffix)) //nolint:gosec // reason: names come from the manifest Export wrote
		if errors.Is(err, fs.ErrNotExist) {
			report.Waiting = append(report.Waiting, label)
			continue
		}
		if err != nil {
			return report, fmt.Errorf("read the translation of %s: %w", label, err)
		}
		if err := importLecture(media, entry, string(translation)); err != nil {
			report.Rejected = append(report.Rejected, label+": "+err.Error())
			continue
		}
		report.Written = append(report.Written, label)
	}
	return report, nil
}

func importLecture(media *os.Root, entry handoffEntry, translation string) error {
	if _, err := media.Stat(entry.TargetPath); err == nil {
		return fmt.Errorf("%s already has Vietnamese subtitles", entry.TargetPath)
	}
	cues, err := subtitle.ParseSRTFile(media.FS(), entry.SourcePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", entry.SourcePath, err)
	}
	translated, err := translate.Accept(cues, translation)
	if err != nil {
		return err
	}
	return translate.WriteSubtitle(media, entry.TargetPath, translated)
}

// matchingCourses picks the course named by title or id, or every course when none is named.
func matchingCourses(courses []library.Course, wanted string) ([]library.Course, error) {
	if wanted == "" {
		return courses, nil
	}
	index := slices.IndexFunc(courses, func(course library.Course) bool { return course.Title == wanted || course.ID == wanted })
	if index < 0 {
		titles := make([]string, len(courses))
		for position, course := range courses {
			titles[position] = course.Title
		}
		return nil, fmt.Errorf("%w %q; courses: %s", ErrCourseNotInLibrary, wanted, strings.Join(titles, ", "))
	}
	return courses[index : index+1], nil
}
