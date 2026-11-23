package library

import (
	"cmp"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	chapterDirName  = regexp.MustCompile(`^(\d{2,}) - (.+)$`)
	lectureFileName = regexp.MustCompile(`^(\d{3}) `)
)

func Scan(downloads fs.FS) (*Library, error) {
	courseDirs, err := visibleEntries(downloads, ".", true)
	if err != nil {
		return nil, err
	}
	var courses []Course
	for _, courseDir := range courseDirs {
		course, err := scanCourse(downloads, courseDir)
		if err != nil {
			return nil, err
		}
		if len(course.Chapters) > 0 {
			courses = append(courses, course)
		}
	}
	return newLibrary(courses)
}

func scanCourse(downloads fs.FS, courseDir string) (Course, error) {
	chapterDirs, err := visibleEntries(downloads, courseDir, true)
	if err != nil {
		return Course{}, err
	}
	course := Course{ID: shortID("course", courseDir), Title: courseDir}
	for _, chapterDir := range chapterDirs {
		match := chapterDirName.FindStringSubmatch(chapterDir)
		if match == nil {
			continue
		}
		index, _ := strconv.Atoi(match[1])
		lectures, err := scanChapter(downloads, path.Join(courseDir, chapterDir))
		if err != nil {
			return Course{}, err
		}
		course.Chapters = append(course.Chapters, Chapter{Index: index, Title: match[2], Lectures: lectures})
	}
	slices.SortFunc(course.Chapters, func(left, right Chapter) int { return cmp.Compare(left.Index, right.Index) })
	return course, nil
}

func scanChapter(downloads fs.FS, chapterDir string) ([]Lecture, error) {
	names, err := visibleEntries(downloads, chapterDir, false)
	if err != nil {
		return nil, err
	}
	groups := map[int][]string{}
	for _, name := range names {
		match := lectureFileName.FindStringSubmatch(name)
		if match == nil {
			continue
		}
		number, _ := strconv.Atoi(match[1])
		groups[number] = append(groups[number], name)
	}

	lectures := make([]Lecture, 0, len(groups))
	for number, groupNames := range groups {
		prefix := fmt.Sprintf("%03d ", number)
		lecture := buildLecture(lectureGroup{chapterDir: chapterDir, prefix: prefix, names: groupNames}, number)
		lecture.Key = path.Join(chapterDir, prefix+lecture.Title)
		lecture.ID = shortID("lecture", lecture.Key)
		for position := range lecture.Attachments {
			lecture.Attachments[position].ID = shortID("attachment", lecture.Attachments[position].Path)
		}
		lectures = append(lectures, lecture)
	}
	slices.SortFunc(lectures, func(left, right Lecture) int { return cmp.Compare(left.Number, right.Number) })
	return lectures, nil
}

func visibleEntries(downloads fs.FS, dir string, wantDirs bool) ([]string, error) {
	entries, err := fs.ReadDir(downloads, dir)
	if err != nil {
		return nil, fmt.Errorf("read library dir %q: %w", dir, err)
	}
	var names []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") || entry.IsDir() != wantDirs {
			continue
		}
		names = append(names, entry.Name())
	}
	return names, nil
}
