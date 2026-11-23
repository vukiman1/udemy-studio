package library

import (
	"encoding/json"
	"os"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const manifestPath = "../../testdata/library/manifest.json"

var cmpIgnoreIdentity = cmpopts.IgnoreFields(Lecture{}, "ID", "Key")

func manifestFS(t *testing.T) fstest.MapFS {
	t.Helper()
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest struct {
		Files []struct {
			Path string `json:"path"`
		} `json:"files"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	tree := fstest.MapFS{}
	for _, file := range manifest.Files {
		tree[file.Path] = &fstest.MapFile{}
	}
	return tree
}

func scanFS(t *testing.T, tree fstest.MapFS) *Library {
	t.Helper()
	scanned, err := Scan(tree)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	return scanned
}

func findLecture(t *testing.T, course Course, chapterIndex, number int) Lecture {
	t.Helper()
	for _, chapter := range course.Chapters {
		if chapter.Index != chapterIndex {
			continue
		}
		for _, lecture := range chapter.Lectures {
			if lecture.Number == number {
				return lecture
			}
		}
	}
	t.Fatalf("lecture %02d/%03d not found", chapterIndex, number)
	return Lecture{}
}

func TestScanRealLibraryShape(t *testing.T) {
	scanned := scanFS(t, manifestFS(t))
	if len(scanned.Courses) != 1 {
		t.Fatalf("courses = %d, want 1", len(scanned.Courses))
	}
	course := scanned.Courses[0]
	if course.Title != "fundamentals-of-backend-communications-and-protocols" {
		t.Errorf("course title = %q", course.Title)
	}
	lectureCount, videoCount, bilingualCount := 0, 0, 0
	for _, chapter := range course.Chapters {
		for _, lecture := range chapter.Lectures {
			lectureCount++
			if lecture.Kind == LectureKindVideo {
				videoCount++
			}
			if len(lecture.Subtitles) == 2 {
				bilingualCount++
			}
		}
	}
	if len(course.Chapters) != 8 || lectureCount != 64 || videoCount != 58 || bilingualCount != 58 {
		t.Errorf("chapters=%d lectures=%d videos=%d bilingual=%d, want 8/64/58/58",
			len(course.Chapters), lectureCount, videoCount, bilingualCount)
	}
}

func TestScanRealLibraryEdgeLectures(t *testing.T) {
	course := scanFS(t, manifestFS(t)).Courses[0]
	const courseDir = "fundamentals-of-backend-communications-and-protocols/"

	t.Run("plain video lecture with bilingual subtitles", func(t *testing.T) {
		lecture := findLecture(t, course, 1, 1)
		want := Lecture{
			Number: 1, Title: "Welcome", Kind: LectureKindVideo,
			VideoPath: courseDir + "01 - Introduction/001 Welcome.mp4",
			Subtitles: []SubtitleFile{
				{Language: "en", Path: courseDir + "01 - Introduction/001 Welcome_en.srt"},
				{Language: "vi", Path: courseDir + "01 - Introduction/001 Welcome_vi.srt"},
			},
		}
		if diff := cmp.Diff(want, lecture, cmpIgnoreIdentity); diff != "" {
			t.Errorf("lecture mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("article that comes with its own video", func(t *testing.T) {
		lecture := findLecture(t, course, 7, 3)
		if lecture.Kind != LectureKindArticle || lecture.Title != "The Journey of a Request to the Backend" {
			t.Errorf("kind=%s title=%q", lecture.Kind, lecture.Title)
		}
		if lecture.VideoPath != courseDir+"07 - Extras/003 journey-request.mp4" {
			t.Errorf("video = %q", lecture.VideoPath)
		}
		if len(lecture.Attachments) != 1 || lecture.Attachments[0].Name != "The-Journey-of-a-Request-to-the-Backend.pdf" {
			t.Errorf("attachments = %+v", lecture.Attachments)
		}
	})

	t.Run("drm lecture with captions but no video", func(t *testing.T) {
		lecture := findLecture(t, course, 8, 5)
		if lecture.Kind != LectureKindUnavailable || lecture.Title != "HTTP Graceful Connection Shutdown" || len(lecture.Subtitles) != 2 {
			t.Errorf("kind=%s title=%q subtitles=%d", lecture.Kind, lecture.Title, len(lecture.Subtitles))
		}
	})

	t.Run("resources article keeps every attachment", func(t *testing.T) {
		lecture := findLecture(t, course, 1, 5)
		if lecture.Kind != LectureKindArticle || len(lecture.Attachments) != 27 {
			t.Errorf("kind=%s attachments=%d, want article with 27", lecture.Kind, len(lecture.Attachments))
		}
	})
}

func TestScanMatchesSubtitlesWhenOnlyTheVideoNameLostItsEmoji(t *testing.T) {
	scanned := scanFS(t, fstest.MapFS{
		"course/01 - Intro/001 Ship it  fast.mp4":     {},
		"course/01 - Intro/001 Ship it 🚀 fast_en.srt": {},
	})
	lecture := findLecture(t, scanned.Courses[0], 1, 1)
	if lecture.Kind != LectureKindVideo || len(lecture.Subtitles) != 1 {
		t.Errorf("kind=%s subtitles=%+v, want video with the emoji-named subtitle", lecture.Kind, lecture.Subtitles)
	}
}

func TestScanIgnoresFilesOutsideTheNamingContract(t *testing.T) {
	scanned := scanFS(t, fstest.MapFS{
		"course/.DS_Store":                  {},
		"course/notes.txt":                  {},
		"course/01 - Intro/.DS_Store":       {},
		"course/01 - Intro/readme.md":       {},
		"course/01 - Intro/001 Welcome.mp4": {},
		"course/extras/001 Stray.mp4":       {},
		".hidden-course/01 - A/001 B.mp4":   {},
	})
	if len(scanned.Courses) != 1 || len(scanned.Courses[0].Chapters) != 1 || len(scanned.Courses[0].Chapters[0].Lectures) != 1 {
		t.Errorf("scanned = %+v, want one course, chapter and lecture", scanned.Courses)
	}
}

func TestScanSortsChaptersAndLecturesNumerically(t *testing.T) {
	course := scanFS(t, fstest.MapFS{
		"course/10 - Ten/002 B.mp4": {},
		"course/10 - Ten/010 C.mp4": {},
		"course/02 - Two/001 A.mp4": {},
		"course/10 - Ten/001 A.mp4": {},
	}).Courses[0]
	var order []string
	for _, chapter := range course.Chapters {
		for _, lecture := range chapter.Lectures {
			order = append(order, chapter.Title+"/"+lecture.Title)
		}
	}
	want := []string{"Two/A", "Ten/A", "Ten/B", "Ten/C"}
	if diff := cmp.Diff(want, order); diff != "" {
		t.Errorf("order mismatch (-want +got):\n%s", diff)
	}
}

func TestLookupsResolveIDsAndRejectUnknownOnes(t *testing.T) {
	scanned := scanFS(t, manifestFS(t))
	lecture := findLecture(t, scanned.Courses[0], 7, 3)

	found, ok := scanned.Lecture(lecture.ID)
	if !ok || found.VideoPath != lecture.VideoPath {
		t.Errorf("Lecture(%s) = %+v, %t", lecture.ID, found, ok)
	}
	attachment, ok := scanned.Attachment(lecture.Attachments[0].ID)
	if !ok || attachment.Path != lecture.Attachments[0].Path {
		t.Errorf("Attachment(%s) = %+v, %t", lecture.Attachments[0].ID, attachment, ok)
	}
	if _, ok := scanned.Course(scanned.Courses[0].ID); !ok {
		t.Error("Course(own id) not found")
	}
	for _, unknown := range []string{"", "../etc/passwd", "zzzzzzzzzzzz"} {
		if _, ok := scanned.Lecture(unknown); ok {
			t.Errorf("Lecture(%q) resolved, want unknown", unknown)
		}
	}
}

func TestIDsAreStableAcrossScans(t *testing.T) {
	first := scanFS(t, manifestFS(t)).Courses[0]
	second := scanFS(t, manifestFS(t)).Courses[0]
	if first.ID != second.ID || findLecture(t, first, 3, 2).ID != findLecture(t, second, 3, 2).ID {
		t.Error("ids changed between scans of the same tree")
	}
}

func TestArticleKeepsItsIdentityWhenItsVideoGainsSubtitles(t *testing.T) {
	before := fstest.MapFS{
		"course/07 - Extras/003 The Journey.html":    {},
		"course/07 - Extras/003 journey-request.mp4": {},
	}
	after := fstest.MapFS{
		"course/07 - Extras/003 The Journey.html":       {},
		"course/07 - Extras/003 journey-request.mp4":    {},
		"course/07 - Extras/003 journey-request_vi.srt": {},
	}
	original := findLecture(t, scanFS(t, before).Courses[0], 7, 3)
	translated := findLecture(t, scanFS(t, after).Courses[0], 7, 3)
	if translated.ID != original.ID || translated.Kind != LectureKindArticle || translated.Title != "The Journey" {
		t.Errorf("after adding subtitles: id %s→%s kind=%s title=%q", original.ID, translated.ID, translated.Kind, translated.Title)
	}
	if translated.VideoPath == "" || len(translated.Subtitles) != 1 {
		t.Errorf("video=%q subtitles=%+v, want the attached video with its subtitle", translated.VideoPath, translated.Subtitles)
	}
}
