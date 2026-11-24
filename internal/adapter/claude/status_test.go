package claude

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

func TestAClaudeKeyBarredFromTheModelIsRejectedLikeAnUnknownOne(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(status)
			_, _ = writer.Write([]byte(`{"type":"error","error":{"type":"permission_error","message":"refused"}}`))
		}))
		_, err := Client{baseURL: server.URL}.Complete(t.Context(), secret.New("key"), "model", llm.Prompt{})
		server.Close()
		if !errors.Is(err, llm.ErrInvalidAPIKey) {
			t.Errorf("Complete on HTTP %d = %v; want a rejected key", status, err)
		}
	}
}
