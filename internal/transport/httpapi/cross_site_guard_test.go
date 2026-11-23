package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCrossSiteWritesAreRejected(t *testing.T) {
	inner := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(http.StatusAccepted) })
	cases := []struct {
		name    string
		method  string
		headers map[string]string
		want    int
	}{
		{name: "text/plain POST from another site", method: http.MethodPost, headers: map[string]string{"Origin": "https://evil.example", "Sec-Fetch-Site": "cross-site"}, want: http.StatusForbidden},
		{name: "POST from another site without fetch metadata", method: http.MethodPost, headers: map[string]string{"Origin": "https://evil.example"}, want: http.StatusForbidden},
		{name: "POST from another local port", method: http.MethodPost, headers: map[string]string{"Origin": "http://127.0.0.1:3000"}, want: http.StatusForbidden},
		{name: "POST from the app itself", method: http.MethodPost, headers: map[string]string{"Origin": "http://127.0.0.1:8765", "Sec-Fetch-Site": "same-origin"}, want: http.StatusAccepted},
		{name: "POST from a command line client", method: http.MethodPost, headers: nil, want: http.StatusAccepted},
		{name: "DELETE typed into the address bar", method: http.MethodDelete, headers: map[string]string{"Sec-Fetch-Site": "none"}, want: http.StatusAccepted},
		{name: "cross-site GET is left to CORS", method: http.MethodGet, headers: map[string]string{"Origin": "https://evil.example", "Sec-Fetch-Site": "cross-site"}, want: http.StatusAccepted},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), testCase.method, "http://127.0.0.1:8765/api/jobs", strings.NewReader("{}"))
			for name, value := range testCase.headers {
				request.Header.Set(name, value)
			}
			recorder := httptest.NewRecorder()
			RejectCrossSiteWrites(inner).ServeHTTP(recorder, request)
			if recorder.Code != testCase.want {
				t.Errorf("status = %d, want %d", recorder.Code, testCase.want)
			}
		})
	}
}
