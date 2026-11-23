package udemyapi

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
)

const testPortalURL = "https://example-portal.udemy.com"

type fakeTransport struct {
	mu        sync.Mutex
	responses map[string][]response
	requested []string
}

func (fake *fakeTransport) reply(target string, status int, body string) {
	if fake.responses == nil {
		fake.responses = map[string][]response{}
	}
	fake.responses[target] = append(fake.responses[target], response{status: status, body: []byte(body)})
}

func (fake *fakeTransport) get(_ context.Context, target string) (response, error) {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	fake.requested = append(fake.requested, target)
	queued := fake.responses[target]
	if len(queued) == 0 {
		return response{status: 404, body: []byte(`{"detail": "Not found."}`)}, nil
	}
	if len(queued) > 1 {
		fake.responses[target] = queued[1:]
	}
	return queued[0], nil
}

func (fake *fakeTransport) requestedMatching(fragment string) int {
	count := 0
	for _, target := range fake.requested {
		if strings.Contains(target, fragment) {
			count++
		}
	}
	return count
}

func newTestClient(t *testing.T, fake *fakeTransport) *Client {
	t.Helper()
	return &Client{transport: fake, portalURL: testPortalURL, logger: slog.New(slog.DiscardHandler)}
}
