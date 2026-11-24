package library

import (
	"testing"
	"testing/fstest"
)

func TestACoursesSizeCountsEveryFileInItsFolderOnly(t *testing.T) {
	downloads := fstest.MapFS{
		"course/01 - Intro/001 Welcome.mp4":    {Data: make([]byte, 1000)},
		"course/01 - Intro/001 Welcome_en.srt": {Data: make([]byte, 20)},
		"course/01 - Intro/slides.pdf":         {Data: make([]byte, 300)},
		"other/01 - Intro/001 Other.mp4":       {Data: make([]byte, 5000)},
	}
	size, err := SizeBytes(downloads, Course{Dir: "course"})
	if err != nil || size != 1320 {
		t.Errorf("SizeBytes = %d, %v; want 1320, the course's own files", size, err)
	}
}
