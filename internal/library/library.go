package library

import (
	"errors"
	"fmt"
)

var ErrDuplicateID = errors.New("library id collision")

type Library struct {
	Courses     []Course
	courses     map[string]Course
	lectures    map[string]Lecture
	attachments map[string]Attachment
	byKey       map[string]lectureLocation
}

type lectureLocation struct {
	courseID, lectureID string
}

func newLibrary(courses []Course) (*Library, error) {
	scanned := &Library{
		Courses:     courses,
		courses:     map[string]Course{},
		lectures:    map[string]Lecture{},
		attachments: map[string]Attachment{},
		byKey:       map[string]lectureLocation{},
	}
	for _, course := range courses {
		if err := insertUnique(scanned.courses, course.ID, course, course.Title); err != nil {
			return nil, err
		}
		for _, chapter := range course.Chapters {
			for _, lecture := range chapter.Lectures {
				if err := insertUnique(scanned.lectures, lecture.ID, lecture, lecture.Key); err != nil {
					return nil, err
				}
				scanned.byKey[lecture.Key] = lectureLocation{courseID: course.ID, lectureID: lecture.ID}
				for _, attachment := range lecture.Attachments {
					if err := insertUnique(scanned.attachments, attachment.ID, attachment, attachment.Path); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return scanned, nil
}

func insertUnique[T any](index map[string]T, id string, value T, label string) error {
	if _, exists := index[id]; exists {
		return fmt.Errorf("%w: %s for %q", ErrDuplicateID, id, label)
	}
	index[id] = value
	return nil
}

func (scanned *Library) Course(id string) (Course, bool) {
	course, ok := scanned.courses[id]
	return course, ok
}

func (scanned *Library) Lecture(id string) (Lecture, bool) {
	lecture, ok := scanned.lectures[id]
	return lecture, ok
}

func (scanned *Library) Attachment(id string) (Attachment, bool) {
	attachment, ok := scanned.attachments[id]
	return attachment, ok
}

// KeyRenames maps the path keys of lectures that now have a Udemy id to that id key.
func (scanned *Library) KeyRenames() map[string]string {
	renames := map[string]string{}
	for _, lecture := range scanned.lectures {
		if lecture.PathKey != "" {
			renames[lecture.PathKey] = lecture.Key
		}
	}
	return renames
}

// LectureByKey finds a lecture and its course by the key its progress and search entries are stored under.
func (scanned *Library) LectureByKey(key string) (Course, Lecture, bool) {
	location, ok := scanned.byKey[key]
	if !ok {
		return Course{}, Lecture{}, false
	}
	return scanned.courses[location.courseID], scanned.lectures[location.lectureID], true
}
