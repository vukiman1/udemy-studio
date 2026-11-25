package httpapi

import (
	"maps"
	"net/http"
	"strings"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/library"
)

func TestVideoSupportsRangeRequests(t *testing.T) {
	app := newTestApp(t)
	result := app.do(t, http.MethodGet, "/media/"+app.lectureID(t, 1), nil, "Range", "bytes=4-7")
	if result.status != http.StatusPartialContent || string(result.body) != "4567" || result.header.Get("Content-Type") != videoContentType {
		t.Errorf("status=%d type=%q body=%q", result.status, result.header.Get("Content-Type"), result.body)
	}
}

func TestLibraryFilesRefuseWhatIsNotInTheIndex(t *testing.T) {
	app := newTestApp(t)
	cases := map[string]string{
		"unknown lecture id":         "/media/nope",
		"encoded traversal as an id": "/media/..%2F..%2Fetc%2Fpasswd",
		"lecture without a video":    "/media/" + app.lectureID(t, 3),
		"symlink pointing outside":   "/media/" + app.lectureID(t, 4),
		"lecture without an article": "/articles/" + app.lectureID(t, 1),
		"unknown attachment":         "/attachments/nope",
	}
	for name, path := range cases {
		t.Run(name, func(t *testing.T) {
			result := app.do(t, http.MethodGet, path, nil)
			assertProblem(t, result, http.StatusNotFound)
			if strings.Contains(string(result.body), secretBytes) {
				t.Errorf("response leaked a file outside the library: %q", result.body)
			}
		})
	}
}

func TestArticleIsServedInASandboxWithoutScripts(t *testing.T) {
	app := newTestApp(t)
	result := app.do(t, http.MethodGet, "/articles/"+app.lectureID(t, 2), nil)
	if csp := result.header.Get("Content-Security-Policy"); result.status != http.StatusOK || !strings.HasPrefix(csp, "sandbox;") {
		t.Errorf("status=%d csp=%q, want a sandbox that allows nothing", result.status, csp)
	}
}

func TestQuizPageRunsItsScriptsInASandbox(t *testing.T) {
	files := maps.Clone(libraryFixture)
	files["course/"+library.ManifestFile] = `{"format": 1, "lectures": [{"udemy_id": 7, "class": "quiz", "chapter_dir": "01 - Intro", "file": "002 Read me.html"}]}`
	app := newTestAppWith(t, files)
	result := app.do(t, http.MethodGet, "/articles/"+app.lectureID(t, 2), nil)
	if csp := result.header.Get("Content-Security-Policy"); result.status != http.StatusOK || !strings.HasPrefix(csp, "sandbox allow-scripts;") || strings.Contains(csp, "allow-same-origin") {
		t.Errorf("status=%d csp=%q, want scripts in an opaque-origin sandbox", result.status, csp)
	}
}

func TestArticleAndQuizPagesWearTheAppsTheme(t *testing.T) {
	page := "<html><head><title>t</title></HEAD><body><h1>article</h1></body></html>"
	cases := map[library.LectureKind]string{library.LectureKindArticle: articleThemeCSS, library.LectureKindQuiz: quizThemeCSS}
	for kind, themeCSS := range cases {
		t.Run(string(kind), func(t *testing.T) {
			files := maps.Clone(libraryFixture)
			files["course/01 - Intro/002 Read me.html"] = page
			files["course/"+library.ManifestFile] = `{"format": 1, "lectures": [{"udemy_id": 7, "class": "` + string(kind) + `", "chapter_dir": "01 - Intro", "file": "002 Read me.html"}]}`
			app := newTestAppWith(t, files)
			body := string(app.do(t, http.MethodGet, "/articles/"+app.lectureID(t, 2), nil).body)
			want := "<title>t</title><style>" + paletteCSS + themeCSS + "</style></HEAD><body><h1>article</h1>"
			if !strings.Contains(body, want) {
				t.Errorf("page = %.200q…, want the %s theme at the end of its head and the page kept", body, kind)
			}
		})
	}
}

func TestAttachmentDownloadsUnderItsOwnName(t *testing.T) {
	app := newTestApp(t)
	lecture, _ := app.library.Lecture(app.lectureID(t, 2))
	result := app.do(t, http.MethodGet, "/attachments/"+lecture.Attachments[0].ID, nil)
	if result.status != http.StatusOK || result.header.Get("Content-Disposition") != `attachment; filename=slides.pdf` || result.header.Get("Content-Type") != "application/pdf" {
		t.Errorf("status=%d disposition=%q type=%q", result.status, result.header.Get("Content-Disposition"), result.header.Get("Content-Type"))
	}
}
