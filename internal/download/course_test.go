package download

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const (
	vttBody          = "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nHello\n"
	childPlaylist    = "#EXTM3U\n#EXT-X-ENDLIST\n"
	failingURL       = "https://cdn.example/fails"
	brokenCaptionURL = "https://cdn.example/broken.vtt"
	testFileMode     = 0o600
	hlsPlaylistURL   = "https://example-portal.udemy.com/assets/77/files/hls/1080.m3u8"
)

var errFetchFailed = errors.New("aria2c exited with 1")

type fakeTools struct {
	mutex      sync.Mutex
	fetched    []string
	hlsOutputs []string
	cancel     context.CancelFunc
}

func (tools *fakeTools) FetchFile(_ context.Context, fileURL, dir, name string) error {
	tools.mutex.Lock()
	defer tools.mutex.Unlock()
	tools.fetched = append(tools.fetched, name)
	if tools.cancel != nil {
		tools.cancel()
		return context.Canceled
	}
	if fileURL == failingURL {
		return errFetchFailed
	}
	content := "fetched " + fileURL
	if strings.HasSuffix(name, vttExtension) && fileURL != brokenCaptionURL {
		content = vttBody
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), testFileMode); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(dir, name) + aria2ControlSuffix); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func (tools *fakeTools) FetchHLS(_ context.Context, playlistURL, outputTemplate string) error {
	parsed, err := url.Parse(playlistURL)
	if err != nil {
		return err
	}
	if playlist, err := os.ReadFile(parsed.Path); err != nil || string(playlist) != childPlaylist {
		return errors.New("yt-dlp did not get the saved child playlist")
	}
	tools.hlsOutputs = append(tools.hlsOutputs, outputTemplate)
	return os.WriteFile(strings.Replace(outputTemplate, ytDLPExtensionTemplate, videoExtension, 1), []byte("hls"), testFileMode)
}

func (tools *fakeTools) FetchPlaylist(context.Context, string) (string, error) {
	return childPlaylist, nil
}

func samplePlan() udemy.Plan {
	videoLecture := udemy.LecturePlan{
		Number: 1, ID: 101, Title: "001 Welcome", Kind: udemy.KindVideo, File: "001 Welcome.mp4",
		Video:    &udemy.VideoChoice{Type: udemy.SourceMP4, URL: "https://cdn.example/welcome.mp4"},
		Captions: []udemy.CaptionPlan{{Language: "en", URL: "https://cdn.example/en.vtt", File: "001 Welcome_en.vtt", SRTFile: "001 Welcome_en.srt"}},
		Assets: []udemy.AssetPlan{
			{Type: udemy.AssetFile, File: "001 slides.pdf", URL: "https://cdn.example/slides.pdf"},
			{Type: udemy.AssetExternalLink, File: "001 Docs.url", URL: "https://docs.example"},
		},
	}
	hlsLecture := udemy.LecturePlan{
		Number: 2, ID: 102, Title: "002 Deep dive", Kind: udemy.KindVideo, File: "002 Deep dive.mp4",
		Video: &udemy.VideoChoice{Type: udemy.SourceHLS, URL: hlsPlaylistURL, PlaylistFile: "index_77_1920x1080.m3u8"},
	}
	article := udemy.LecturePlan{
		Number: 3, ID: 103, Title: "003 Notes 🚀", Kind: udemy.KindHTML, File: "003 Notes 🚀.html", HTMLBody: "<p>body</p>",
		Assets: []udemy.AssetPlan{{Type: udemy.AssetArticle, File: "003 Notes 🚀.html"}},
	}
	drm := udemy.LecturePlan{Number: 4, Title: "004 Locked", Kind: udemy.KindDRM, File: "004 Locked.mp4"}
	quiz := udemy.LecturePlan{Number: 5, Title: "005 Quiz", Kind: udemy.KindQuiz, File: "005 Quiz.html"}
	return udemy.Plan{CourseName: "backend-basics", CourseDir: "backend-basics", Chapters: []udemy.ChapterPlan{
		{Index: 1, Dir: "01 - Intro", Lectures: []udemy.LecturePlan{videoLecture, hlsLecture, article, drm, quiz}},
		{Index: 2, Dir: "02 - Skipped", Lectures: []udemy.LecturePlan{{Number: 1, Title: "001 Other", Kind: udemy.KindHTML, File: "001 Other.html", HTMLBody: "x"}}},
	}}
}

func runCourse(t *testing.T, ctx context.Context, plan udemy.Plan, tools *fakeTools, outputDir string) (Summary, []jobs.Event, error) {
	t.Helper()
	var events []jobs.Event
	options := Options{OutputDir: outputDir, TempDir: t.TempDir(), Chapters: ChapterFilter{{first: 1, last: 1}}}
	summary, err := Course(ctx, plan, Tools{Files: tools, Streams: tools, Playlists: tools}, options, slog.New(slog.DiscardHandler), func(event jobs.Event) {
		events = append(events, event)
	})
	return summary, events, err
}

func treeOf(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, _ := filepath.Rel(root, path)
		files = append(files, relative)
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	return files
}

func TestCourseWritesThePythonLayoutAndSkipsItOnARerun(t *testing.T) {
	outputDir := t.TempDir()
	tools := &fakeTools{}
	summary, events, err := runCourse(t, t.Context(), samplePlan(), tools, outputDir)
	if err != nil {
		t.Fatalf("Course: %v", err)
	}
	wantTree := []string{
		"backend-basics/.course.json",
		"backend-basics/01 - Intro/001 Docs.url",
		"backend-basics/01 - Intro/001 Welcome.mp4",
		"backend-basics/01 - Intro/001 Welcome_en.srt",
		"backend-basics/01 - Intro/001 slides.pdf",
		"backend-basics/01 - Intro/002 Deep dive.mp4",
		"backend-basics/01 - Intro/003 Notes 🚀.html",
		"backend-basics/01 - Intro/external-links.txt",
	}
	if diff := cmp.Diff(wantTree, treeOf(t, outputDir)); diff != "" {
		t.Errorf("tree (-want +got):\n%s", diff)
	}
	wantSummary := Summary{Downloaded: 3, Failed: []string{"004 Locked"}}
	if diff := cmp.Diff(wantSummary, summary); diff != "" {
		t.Errorf("summary (-want +got):\n%s", diff)
	}
	if events[0].TotalLectures != 4 || events[len(events)-1].Kind != jobs.EventSummary {
		t.Errorf("events = %+v", events)
	}
	wantHLSOutput := filepath.Join(outputDir, "backend-basics", "01 - Intro", "002 Deep dive.%(ext)s")
	if !slices.Equal(tools.hlsOutputs, []string{wantHLSOutput}) {
		t.Errorf("yt-dlp output = %v, want %s", tools.hlsOutputs, wantHLSOutput)
	}

	rerun := &fakeTools{}
	summary, _, err = runCourse(t, t.Context(), samplePlan(), rerun, outputDir)
	if err != nil {
		t.Fatalf("rerun: %v", err)
	}
	// Attachments are always fetched again, as in Python; aria2c's -c keeps what is there.
	wantRerun := Summary{AlreadyPresent: 3, Failed: []string{"004 Locked"}}
	if diff := cmp.Diff(wantRerun, summary); diff != "" || !slices.Equal(rerun.fetched, []string{"001 slides.pdf"}) {
		t.Errorf("rerun summary (-want +got):\n%s\nfetched = %v", diff, rerun.fetched)
	}
	links, _ := fs.ReadFile(os.DirFS(outputDir), "backend-basics/01 - Intro/"+externalLinksFile)
	if string(links) != "\n001 Docs.url\nhttps://docs.example\n" {
		t.Errorf("external links = %q, want one entry", links)
	}
}

func TestFailedCaptionIsRetriedFourTimesThenRecorded(t *testing.T) {
	plan := samplePlan()
	lecture := &plan.Chapters[0].Lectures[0]
	lecture.Captions[0].URL = failingURL
	lecture.Assets = nil
	plan.Chapters[0].Lectures = plan.Chapters[0].Lectures[:1]
	tools := &fakeTools{}
	summary, events, err := runCourse(t, t.Context(), plan, tools, t.TempDir())
	if err != nil {
		t.Fatalf("Course: %v", err)
	}
	attempts := 0
	for _, name := range tools.fetched {
		if name == lecture.Captions[0].File {
			attempts++
		}
	}
	if attempts != captionAttempts || !slices.Equal(summary.Failed, []string{"001 Welcome (caption: en)"}) {
		t.Errorf("attempts = %d, failed = %v", attempts, summary.Failed)
	}
	if lectureEnd := events[len(events)-2]; lectureEnd.Kind != jobs.EventLectureFailed {
		t.Errorf("lecture ended with %+v, want a failure", lectureEnd)
	}
}

func TestCancelStopsTheCourseAtTheRunningDownload(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	tools := &fakeTools{cancel: cancel}
	_, _, err := runCourse(t, ctx, samplePlan(), tools, t.TempDir())
	if !errors.Is(err, context.Canceled) || len(tools.fetched) != 1 {
		t.Errorf("err = %v after fetching %v, want a cancel after the first download", err, tools.fetched)
	}
}

func TestChapterFilterReadsListsAndRanges(t *testing.T) {
	filter, err := ParseChapterFilter("1, 3-5,9")
	if err != nil {
		t.Fatalf("ParseChapterFilter: %v", err)
	}
	var included []int
	for index := range 11 {
		if filter.includes(index) {
			included = append(included, index)
		}
	}
	if !slices.Equal(included, []int{1, 3, 4, 5, 9}) {
		t.Errorf("included chapters = %v", included)
	}
	for _, invalid := range []string{"5-3", "x", "1-2-3", "-1"} {
		if _, err := ParseChapterFilter(invalid); !errors.Is(err, ErrInvalidChapterFilter) {
			t.Errorf("ParseChapterFilter(%q) error = %v", invalid, err)
		}
	}
}

func TestLibraryKeysDownloadedLecturesByTheirUdemyID(t *testing.T) {
	outputDir := t.TempDir()
	if _, _, err := runCourse(t, t.Context(), samplePlan(), &fakeTools{}, outputDir); err != nil {
		t.Fatalf("Course: %v", err)
	}
	scanned, err := library.Scan(os.DirFS(outputDir))
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	keys := map[string]string{}
	for _, lecture := range scanned.Courses[0].Chapters[0].Lectures {
		keys[lecture.Title] = lecture.Key
	}
	want := map[string]string{
		"Welcome":   "udemy-lecture:101",
		"Deep dive": "udemy-lecture:102",
		"Notes 🚀":   "udemy-lecture:103",
	}
	if diff := cmp.Diff(want, keys); diff != "" {
		t.Errorf("lecture keys (-want +got):\n%s", diff)
	}
	renames := scanned.KeyRenames()
	if renames["backend-basics/01 - Intro/001 Welcome"] != "udemy-lecture:101" {
		t.Errorf("KeyRenames = %v, want the path key of Welcome moved to its id", renames)
	}
}

func TestFileAria2cLeftPartialIsDownloadedAgain(t *testing.T) {
	outputDir := t.TempDir()
	chapterDir := filepath.Join(outputDir, "backend-basics", "01 - Intro")
	if err := os.MkdirAll(chapterDir, dirMode); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, name := range []string{"001 Welcome.mp4", "001 Welcome.mp4.aria2", "001 Welcome_en.vtt", "001 Welcome_en.vtt.aria2"} {
		if err := os.WriteFile(filepath.Join(chapterDir, name), []byte("partial"), testFileMode); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	tools := &fakeTools{}
	if _, _, err := runCourse(t, t.Context(), samplePlan(), tools, outputDir); err != nil {
		t.Fatalf("Course: %v", err)
	}
	for _, name := range []string{"001 Welcome.mp4", "001 Welcome_en.vtt"} {
		if !slices.Contains(tools.fetched, name) {
			t.Errorf("%s was left partial, fetched only %v", name, tools.fetched)
		}
	}
}

func TestCaptionThatDoesNotConvertIsRetriedOnTheNextRun(t *testing.T) {
	plan := samplePlan()
	plan.Chapters[0].Lectures = plan.Chapters[0].Lectures[:1]
	plan.Chapters[0].Lectures[0].Captions[0].URL = brokenCaptionURL
	outputDir := t.TempDir()
	summary, _, err := runCourse(t, t.Context(), plan, &fakeTools{}, outputDir)
	if err != nil || !slices.Equal(summary.Failed, []string{"001 Welcome (caption: en)"}) {
		t.Fatalf("summary = %+v, %v; want the caption failed", summary, err)
	}
	rerun := &fakeTools{}
	if _, _, err := runCourse(t, t.Context(), plan, rerun, outputDir); err != nil {
		t.Fatalf("rerun: %v", err)
	}
	if !slices.Contains(rerun.fetched, "001 Welcome_en.vtt") {
		t.Errorf("rerun fetched %v, want the caption downloaded again", rerun.fetched)
	}
}

func TestArticlePageIsWrittenOnceAndAFailedWriteIsReported(t *testing.T) {
	outputDir := t.TempDir()
	pagePath := filepath.Join(outputDir, "backend-basics", "01 - Intro", "003 Notes 🚀.html")
	if _, _, err := runCourse(t, t.Context(), samplePlan(), &fakeTools{}, outputDir); err != nil {
		t.Fatalf("Course: %v", err)
	}
	written, err := fs.ReadFile(os.DirFS(filepath.Dir(pagePath)), filepath.Base(pagePath))
	if err != nil || !strings.Contains(string(written), "Notes 🚀") || !strings.Contains(string(written), "<p>body</p>") {
		t.Fatalf("page = %q, %v; want the article template with title and body", written, err)
	}
	if err := os.WriteFile(pagePath, []byte("kept"), testFileMode); err != nil {
		t.Fatalf("edit page: %v", err)
	}
	if _, _, err := runCourse(t, t.Context(), samplePlan(), &fakeTools{}, outputDir); err != nil {
		t.Fatalf("rerun: %v", err)
	}
	if kept, _ := fs.ReadFile(os.DirFS(filepath.Dir(pagePath)), filepath.Base(pagePath)); string(kept) != "kept" {
		t.Errorf("rerun rewrote the page: %q", kept)
	}

	blocked := t.TempDir()
	if err := os.MkdirAll(filepath.Join(blocked, "backend-basics", "01 - Intro", "003 Notes 🚀.html", "in-the-way"), dirMode); err != nil {
		t.Fatalf("block page path: %v", err)
	}
	summary, _, err := runCourse(t, t.Context(), samplePlan(), &fakeTools{}, blocked)
	if err != nil || !slices.Contains(summary.Failed, "003 Notes 🚀") {
		t.Errorf("failed = %v, %v; want the page that could not be written", summary.Failed, err)
	}
}

func TestYTDLPOutputKeepsAChapterNamedAfterMP4Files(t *testing.T) {
	plan := samplePlan()
	plan.Chapters[0].Dir = "01 - Encoding .mp4 files"
	plan.Chapters[0].Lectures = plan.Chapters[0].Lectures[1:2]
	outputDir := t.TempDir()
	tools := &fakeTools{}
	if _, _, err := runCourse(t, t.Context(), plan, tools, outputDir); err != nil {
		t.Fatalf("Course: %v", err)
	}
	want := filepath.Join(outputDir, "backend-basics", "01 - Encoding .mp4 files", "002 Deep dive.%(ext)s")
	if !slices.Equal(tools.hlsOutputs, []string{want}) {
		t.Errorf("yt-dlp output = %v, want %s", tools.hlsOutputs, want)
	}
}

func TestCaptionLeftWithoutItsSRTIsConvertedWithoutDownloading(t *testing.T) {
	plan := samplePlan()
	plan.Chapters[0].Lectures = plan.Chapters[0].Lectures[:1]
	plan.Chapters[0].Lectures[0].Assets = nil
	outputDir := t.TempDir()
	chapterDir := filepath.Join(outputDir, "backend-basics", "01 - Intro")
	if err := os.MkdirAll(chapterDir, dirMode); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chapterDir, "001 Welcome_en.vtt"), []byte(vttBody), testFileMode); err != nil {
		t.Fatalf("write vtt: %v", err)
	}
	tools := &fakeTools{}
	if _, _, err := runCourse(t, t.Context(), plan, tools, outputDir); err != nil {
		t.Fatalf("Course: %v", err)
	}
	if slices.Contains(tools.fetched, "001 Welcome_en.vtt") || !isFile(filepath.Join(chapterDir, "001 Welcome_en.srt")) {
		t.Errorf("fetched %v; want the existing VTT converted to SRT without a download", tools.fetched)
	}
}
