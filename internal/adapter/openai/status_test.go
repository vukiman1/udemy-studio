package openai

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

func answering(t *testing.T, status int, retryAfter string) Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if retryAfter != "" {
			writer.Header().Set("Retry-After", retryAfter)
		}
		writer.WriteHeader(status)
		_, _ = writer.Write([]byte(`{"error":{"message":"refused"}}`))
	}))
	t.Cleanup(server.Close)
	return Client{http: server.Client(), baseURL: server.URL}
}

func TestHowOpenAIRefusalsEndARun(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		retryAfter string
		want       error
	}{
		{"an unknown key is rejected", http.StatusUnauthorized, "", llm.ErrInvalidAPIKey},
		{"a key barred from the model is rejected", http.StatusForbidden, "", llm.ErrInvalidAPIKey},
		{"a short wait is worth another try", http.StatusTooManyRequests, "2", llm.ErrTemporary},
		{"a long wait means the quota is spent", http.StatusTooManyRequests, "3600", llm.ErrQuotaExhausted},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := answering(t, test.status, test.retryAfter).Complete(t.Context(), secret.New("key"), "model", llm.Prompt{})
			if !errors.Is(err, test.want) {
				t.Errorf("Complete = %v; want %v", err, test.want)
			}
		})
	}
}
