package gemini

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

func answering(t *testing.T, status int) Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(status)
		_, _ = writer.Write([]byte(`{"error":{"message":"refused"}}`))
	}))
	t.Cleanup(server.Close)
	return Client{http: server.Client(), baseURL: server.URL}
}

func TestAForbiddenModelLetsTheNextOneTryWhileAForbiddenKeyIsRejected(t *testing.T) {
	client := answering(t, http.StatusForbidden)
	_, err := client.Complete(t.Context(), secret.New("key"), "model", llm.Prompt{})
	if err == nil || errors.Is(err, llm.ErrInvalidAPIKey) {
		t.Errorf("Complete on 403 = %v; want a per-model failure, not a rejected key", err)
	}
	if err := client.CheckKey(t.Context(), secret.New("key")); !errors.Is(err, llm.ErrInvalidAPIKey) {
		t.Errorf("CheckKey on 403 = %v; want a rejected key", err)
	}
	if _, err := answering(t, http.StatusUnauthorized).Complete(t.Context(), secret.New("key"), "model", llm.Prompt{}); !errors.Is(err, llm.ErrInvalidAPIKey) {
		t.Errorf("Complete on 401 = %v; want a rejected key", err)
	}
}
