package httpapi

import (
	"net/http"
	"strings"
	"testing"
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

func TestArticleIsServedInASandbox(t *testing.T) {
	app := newTestApp(t)
	result := app.do(t, http.MethodGet, "/articles/"+app.lectureID(t, 2), nil)
	if result.status != http.StatusOK || !strings.HasPrefix(result.header.Get("Content-Security-Policy"), "sandbox") {
		t.Errorf("status=%d csp=%q", result.status, result.header.Get("Content-Security-Policy"))
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
