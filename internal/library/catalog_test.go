package library

import (
	"testing"
	"testing/fstest"
)

func TestRescanPicksUpNewlyDownloadedLectures(t *testing.T) {
	tree := fstest.MapFS{"course/01 - Intro/001 Welcome.mp4": {}}
	catalog, err := NewCatalog(tree)
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	before := catalog.Current()
	tree["course/01 - Intro/002 Next.mp4"] = &fstest.MapFile{}
	if err := catalog.Rescan(); err != nil {
		t.Fatalf("Rescan: %v", err)
	}
	if got := len(catalog.Current().Courses[0].Chapters[0].Lectures); got != 2 {
		t.Errorf("lectures after rescan = %d, want 2", got)
	}
	if len(before.Courses[0].Chapters[0].Lectures) != 1 {
		t.Error("an earlier snapshot changed under its readers")
	}
}
