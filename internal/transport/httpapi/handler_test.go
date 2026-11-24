package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/notes"
	"github.com/vukiman1/udemy-studio/internal/notes/notestest"
	"github.com/vukiman1/udemy-studio/internal/progress"
	"github.com/vukiman1/udemy-studio/internal/progress/progresstest"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

const (
	testVersion   = "v1.2.3"
	indexMarkup   = "<!doctype html><title>app</title>"
	scriptContent = "console.log('app')"
	videoBytes    = "0123456789abcdef"
	secretBytes   = "outside the library"
)

const (
	englishSRT    = "1\n00:00:00,000 --> 00:00:01,500\nHello\n\n2\n00:00:01,500 --> 00:00:03,000\nWorld\n"
	vietnameseSRT = "1\n00:00:00,000 --> 00:00:01,500\nXin chào\n\n2\n00:00:01,500 --> 00:00:03,000\nThế giới\n"
)

var libraryFixture = map[string]string{
	"course/01 - Intro/001 Welcome.mp4":        videoBytes,
	"course/01 - Intro/001 Welcome_en.srt":     englishSRT,
	"course/01 - Intro/001 Welcome_vi.srt":     vietnameseSRT,
	"course/01 - Intro/002 Read me.html":       "<h1>article</h1><script>alert(1)</script>",
	"course/01 - Intro/002 slides.pdf":         "%PDF-1.4",
	"course/01 - Intro/003 Drm lecture_en.srt": englishSRT,
}

type testApp struct {
	server  *httptest.Server
	library *library.Library
}

func newTestApp(t *testing.T) testApp {
	t.Helper()
	return newTestAppWith(t, libraryFixture)
}

func newTestAppWith(t *testing.T, files map[string]string) testApp {
	t.Helper()
	return newTestAppWithDeps(t, files, func(*Deps) {})
}

func newTestAppWithDeps(t *testing.T, files map[string]string, customize func(*Deps)) testApp {
	t.Helper()
	root := t.TempDir()
	for relativePath, content := range files {
		writeFile(t, filepath.Join(root, relativePath), content)
	}
	outside := filepath.Join(t.TempDir(), "secret.mp4")
	writeFile(t, outside, secretBytes)
	if err := os.Symlink(outside, filepath.Join(root, "course/01 - Intro/004 Escape.mp4")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	media, err := os.OpenRoot(root)
	if err != nil {
		t.Fatalf("open root: %v", err)
	}
	t.Cleanup(func() { _ = media.Close() })
	catalog, err := library.NewCatalog(media.FS())
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	now := func() time.Time { return time.Date(2026, 11, 23, 8, 0, 0, 0, time.UTC) }
	deps := Deps{
		Version:  testVersion,
		Frontend: fstest.MapFS{"index.html": {Data: []byte(indexMarkup)}, "assets/app-1.js": {Data: []byte(scriptContent)}},
		Catalog:  catalog,
		Media:    media.FS(),
		Progress: progress.NewService(progresstest.NewMemoryStore(), now),
		Notes:    notes.NewService(notestest.NewMemoryStore(), now),
		Logger:   slog.New(slog.DiscardHandler),
	}
	customize(&deps)
	server := httptest.NewServer(NewHandler(deps))
	t.Cleanup(server.Close)
	return testApp{server: server, library: catalog.Current()}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func (app testApp) lectureID(t *testing.T, number int) string {
	t.Helper()
	for _, lecture := range app.library.Courses[0].Chapters[0].Lectures {
		if lecture.Number == number {
			return lecture.ID
		}
	}
	t.Fatalf("no lecture %03d", number)
	return ""
}

type response struct {
	status int
	header http.Header
	body   []byte
}

func (app testApp) do(t *testing.T, method, path string, body any, headers ...string) response {
	t.Helper()
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode body: %v", err)
		}
		payload = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(t.Context(), method, app.server.URL+path, payload)
	if err != nil {
		t.Fatalf("build %s %s: %v", method, path, err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	for index := 0; index+1 < len(headers); index += 2 {
		request.Header.Set(headers[index], headers[index+1])
	}
	result, err := app.server.Client().Do(request)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer result.Body.Close()
	content, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("read body of %s: %v", path, err)
	}
	return response{status: result.StatusCode, header: result.Header, body: content}
}

func decode[T any](t *testing.T, result response) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(result.body, &value); err != nil {
		t.Fatalf("decode %q: %v", result.body, err)
	}
	return value
}

func assertProblem(t *testing.T, result response, wantStatus int) {
	t.Helper()
	if result.status != wantStatus || result.header.Get("Content-Type") != problemContentType {
		t.Errorf("status=%d content-type=%q body=%s, want %d problem+json", result.status, result.header.Get("Content-Type"), result.body, wantStatus)
	}
}

func TestHealthReportsBuildVersion(t *testing.T) {
	health := decode[gen.Health](t, newTestApp(t).do(t, http.MethodGet, "/healthz", nil))
	if health.Status != gen.Ok || health.Version != testVersion {
		t.Errorf("health = %+v", health)
	}
}

func TestFrontendRoutes(t *testing.T) {
	cases := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{name: "root serves index", path: "/", wantStatus: http.StatusOK, wantBody: indexMarkup},
		{name: "existing asset served as is", path: "/assets/app-1.js", wantStatus: http.StatusOK, wantBody: scriptContent},
		{name: "client route falls back to index", path: "/courses/abc/lectures/1", wantStatus: http.StatusOK, wantBody: indexMarkup},
		{name: "asset directory is not listed", path: "/assets/", wantStatus: http.StatusOK, wantBody: indexMarkup},
		{name: "unknown api path is not swallowed by the spa", path: "/api/nope", wantStatus: http.StatusNotFound},
	}
	app := newTestApp(t)
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result := app.do(t, http.MethodGet, testCase.path, nil)
			if result.status != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", result.status, testCase.wantStatus)
			}
			if testCase.wantBody != "" && string(result.body) != testCase.wantBody {
				t.Errorf("body = %q, want %q", result.body, testCase.wantBody)
			}
		})
	}
}

func TestAPIOnlyModeServesNoFrontend(t *testing.T) {
	server := httptest.NewServer(NewHandler(Deps{Version: testVersion, Frontend: nil, Logger: slog.New(slog.DiscardHandler)}))
	t.Cleanup(server.Close)
	app := testApp{server: server}
	if result := app.do(t, http.MethodGet, "/", nil); result.status != http.StatusNotFound {
		t.Errorf("GET / status = %d, want 404 without a built frontend", result.status)
	}
	if result := app.do(t, http.MethodGet, "/healthz", nil); result.status != http.StatusOK {
		t.Errorf("GET /healthz status = %d, want 200", result.status)
	}
}
