package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHostGuardRejectsNamesThatCouldBeRebound(t *testing.T) {
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(http.StatusNoContent) })
	cases := map[string]int{
		"127.0.0.1:8765":         http.StatusNoContent,
		"localhost:8765":         http.StatusNoContent,
		"[::1]:8765":             http.StatusNoContent,
		"192.168.31.106:8765":    http.StatusNoContent,
		"192.168.31.106":         http.StatusNoContent,
		"evil.example:8765":      http.StatusForbidden,
		"localhost.evil.example": http.StatusForbidden,
	}
	for host, want := range cases {
		t.Run(host, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/courses", nil)
			request.Host = host
			recorder := httptest.NewRecorder()
			RejectForeignHosts(inner).ServeHTTP(recorder, request)
			if recorder.Code != want {
				t.Errorf("status = %d, want %d", recorder.Code, want)
			}
		})
	}
}
