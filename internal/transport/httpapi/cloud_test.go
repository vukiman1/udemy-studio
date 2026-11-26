package httpapi

import (
	"context"
	"net/http"
	"testing"
)

const signedLinkPrefix = "https://bucket.example/"

type fakeLinks struct{}

func (fakeLinks) Link(_ context.Context, filePath, _ string) (string, error) {
	return signedLinkPrefix + filePath, nil
}

func newCloudTestApp(t *testing.T) testApp {
	t.Helper()
	return newTestAppWithDeps(t, libraryFixture, func(deps *Deps) {
		deps.IsCloud = true
		deps.MediaLinks = fakeLinks{}
	})
}

func TestCloudVideosRedirectToASignedLink(t *testing.T) {
	app := newCloudTestApp(t)
	client := *app.server.Client()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, app.server.URL+"/media/"+app.lectureID(t, 1), nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	result, err := client.Do(request)
	if err != nil {
		t.Fatalf("GET video: %v", err)
	}
	defer func() { _ = result.Body.Close() }()
	if result.StatusCode != http.StatusFound || result.Header.Get("Location") != signedLinkPrefix+"course/01 - Intro/001 Welcome.mp4" {
		t.Errorf("status=%d location=%q, want a redirect to the video's signed link", result.StatusCode, result.Header.Get("Location"))
	}
}

func TestTheCloudCopyRefusesLibraryOnlyRoutesButServesTheCourse(t *testing.T) {
	app := newCloudTestApp(t)
	for _, path := range []string{"/api/library/tools", "/api/publish", "/api/courses/" + app.library.Courses[0].ID + "/translation/plan"} {
		assertProblem(t, app.do(t, http.MethodGet, path, nil), http.StatusForbidden)
	}
	if result := app.do(t, http.MethodGet, "/api/courses", nil); result.status != http.StatusOK {
		t.Errorf("GET /api/courses status = %d, want 200", result.status)
	}
}
