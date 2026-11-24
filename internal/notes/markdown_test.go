package notes

import (
	"strings"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/library"
)

func TestMarkdownListsNotesInCourseOrderUnderTheirLectures(t *testing.T) {
	course := library.Course{Title: "Backend", Chapters: []library.Chapter{
		{Index: 1, Title: "Intro", Lectures: []library.Lecture{{Key: "k1", Number: 1, Title: "Welcome", VideoPath: "a.mp4"}}},
		{Index: 2, Title: "Protocols", Lectures: []library.Lecture{
			{Key: "k2", Number: 3, Title: "TCP", VideoPath: "b.mp4"},
			{Key: "k3", Number: 4, Title: "Reading", ArticlePath: "c.html"},
		}},
	}}
	byLecture := map[string][]Note{
		"k2": {{AtSeconds: 75.9, Body: "Bắt tay ba bước\nSYN, SYN-ACK, ACK"}, {AtSeconds: 3725, Body: "Head-of-line"}},
		"k3": {{Body: "Đọc lại"}},
	}
	var document strings.Builder
	if err := WriteMarkdown(&document, course, byLecture); err != nil {
		t.Fatalf("WriteMarkdown: %v", err)
	}
	want := `# Backend

## 02 Protocols

### 003 TCP

- **01:15** Bắt tay ba bước
  SYN, SYN-ACK, ACK
- **1:02:05** Head-of-line

### 004 Reading

- Đọc lại
`
	if document.String() != want {
		t.Errorf("markdown =\n%s\nwant\n%s", document.String(), want)
	}
}
