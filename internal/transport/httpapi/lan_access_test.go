package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	testAccessCode = "ABCDEFGHJK"
	lanClientAddr  = "192.168.1.20:52000"
)

func lanRequest(t *testing.T, target string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	passthrough := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { _, _ = writer.Write([]byte("inside")) })
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	request.RemoteAddr = lanClientAddr
	if cookie != nil {
		request.AddCookie(cookie)
	}
	recorder := httptest.NewRecorder()
	RequireLANAccess(testAccessCode, passthrough).ServeHTTP(recorder, request)
	return recorder
}

func TestLANAccessLetsTheHostMachineThrough(t *testing.T) {
	passthrough := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { _, _ = writer.Write([]byte("inside")) })
	for _, remote := range []string{"127.0.0.1:1234", "[::1]:1234"} {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/courses", nil)
		request.RemoteAddr = remote
		recorder := httptest.NewRecorder()
		RequireLANAccess(testAccessCode, passthrough).ServeHTTP(recorder, request)
		if recorder.Body.String() != "inside" {
			t.Errorf("%s: body = %q, want passthrough", remote, recorder.Body.String())
		}
	}
}

func TestLANAccessBlocksDevicesWithoutTheCode(t *testing.T) {
	t.Run("data routes answer with a problem", func(t *testing.T) {
		recorder := lanRequest(t, "/media/abc", nil)
		if recorder.Code != http.StatusUnauthorized || recorder.Header().Get("Content-Type") != problemContentType {
			t.Errorf("status=%d type=%q", recorder.Code, recorder.Header().Get("Content-Type"))
		}
	})
	t.Run("pages show the code form", func(t *testing.T) {
		recorder := lanRequest(t, "/courses/abc", nil)
		if recorder.Code != http.StatusUnauthorized || !strings.Contains(recorder.Body.String(), `name="code"`) {
			t.Errorf("status=%d body=%q", recorder.Code, recorder.Body.String())
		}
	})
	t.Run("a wrong code is not accepted", func(t *testing.T) {
		if recorder := lanRequest(t, "/?code=WRONGCODE1", nil); recorder.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", recorder.Code)
		}
	})
	t.Run("a forged cookie is not accepted", func(t *testing.T) {
		forged := &http.Cookie{Name: lanCookieName, Value: lanCookieValue("guess")} //nolint:gosec // reason: request-side cookie, only name and value are sent
		if recorder := lanRequest(t, "/api/courses", forged); recorder.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", recorder.Code)
		}
	})
}

func TestLANAccessCodeIsExchangedForACookie(t *testing.T) {
	recorder := lanRequest(t, "/courses/abc?code="+testAccessCode+"&tab=notes", nil)
	if recorder.Code != http.StatusSeeOther || recorder.Header().Get("Location") != "/courses/abc?tab=notes" {
		t.Fatalf("status=%d location=%q, want redirect without the code", recorder.Code, recorder.Header().Get("Location"))
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("cookies = %+v, want one HttpOnly SameSite=Lax cookie", cookies)
	}
	if follow := lanRequest(t, "/api/courses", cookies[0]); follow.Body.String() != "inside" {
		t.Errorf("request with the granted cookie = %d %q", follow.Code, follow.Body.String())
	}
}

func TestLANAccessRedirectNeverLeavesTheServer(t *testing.T) {
	recorder := lanRequest(t, "//evil.example/steal?code="+testAccessCode, nil)
	if location := recorder.Header().Get("Location"); location != "/evil.example/steal" {
		t.Errorf("location = %q, want a same-origin path", location)
	}
}
