package udemyapi

import (
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

func TestTLSTransportSendsLoginCookieAndAppHeaders(t *testing.T) {
	var gotCookie, gotClientID, gotUserAgent string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if cookie, err := request.Cookie(accessTokenCookie); err == nil {
			gotCookie = cookie.Value
		}
		gotClientID = request.Header.Get("x-udemy-client-id")
		gotUserAgent = request.UserAgent()
		writer.Header().Set("content-encoding", "gzip")
		compressed := gzip.NewWriter(writer)
		_, _ = compressed.Write([]byte(`{"ok": true}`))
		_ = compressed.Close()
	}))
	defer server.Close()

	transport, err := newTLSTransport(server.URL, secret.New("token-value"))
	if err != nil {
		t.Fatalf("newTLSTransport: %v", err)
	}
	reply, err := transport.get(t.Context(), server.URL+"/api-2.0/x")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if gotCookie != "token-value" || gotClientID != appHeaders["x-udemy-client-id"] || gotUserAgent != appHeaders["user-agent"] || string(reply.body) != `{"ok": true}` {
		t.Errorf("cookie = %q, client id = %q, user agent = %q, body = %q", gotCookie, gotClientID, gotUserAgent, reply.body)
	}
}
