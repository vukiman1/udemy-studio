package udemyapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	maxAttempts       = 10
	retryDelay        = 800 * time.Millisecond
	portalURLTemplate = "https://%s.udemy.com"
	visitPath         = "/api-2.0/visits/current/?fields%5Bvisit%5D=@default,visitor,country&locale=en_US"
)

var (
	ErrChallenged      = errors.New("cloudflare challenged the request: Udemy no longer accepts this client's fingerprint")
	challengeMarkers   = [][]byte{[]byte("challenge-platform"), []byte("<title>Just a moment...</title>")}
	retryableStatuses  = map[int]bool{http.StatusBadGateway: true, http.StatusServiceUnavailable: true}
	authFailedStatuses = map[int]bool{http.StatusUnauthorized: true, http.StatusForbidden: true}
)

type StatusError struct {
	URL    string
	Status int
	Detail string
}

func (err *StatusError) Error() string {
	message := fmt.Sprintf("udemy returned HTTP %d for %s", err.Status, err.URL)
	if err.Detail != "" {
		message += ": " + err.Detail
	}
	if authFailedStatuses[err.Status] {
		message += " (check that the Chrome profile is logged in and enrolled in this course)"
	}
	return message
}

type UnexpectedResponseError struct {
	URL    string
	Status int
	Err    error
}

func (err *UnexpectedResponseError) Error() string {
	var syntaxErr *json.SyntaxError
	if errors.As(err.Err, &syntaxErr) {
		return fmt.Sprintf("udemy returned a non-JSON response (HTTP %d) for %s: likely a login, captcha or block page", err.Status, err.URL)
	}
	return fmt.Sprintf("udemy returned JSON of an unexpected shape (HTTP %d) for %s: %v", err.Status, err.URL, err.Err)
}

func (err *UnexpectedResponseError) Unwrap() error { return err.Err }

type Client struct {
	transport  transport
	portalURL  string
	retryDelay time.Duration
	logger     *slog.Logger
}

func New(portal string, token secret.Secret, logger *slog.Logger) (*Client, error) {
	portalURL := fmt.Sprintf(portalURLTemplate, portal)
	transport, err := newTLSTransport(portalURL, token)
	if err != nil {
		return nil, err
	}
	return &Client{transport: transport, portalURL: portalURL, retryDelay: retryDelay, logger: logger}, nil
}

// Visit picks up Cloudflare's cookies before the API calls and fails fast when the fingerprint is rejected.
func (client *Client) Visit(ctx context.Context) error {
	reply, err := client.getWithRetry(ctx, client.portalURL+visitPath)
	if err != nil {
		return fmt.Errorf("visit %s: %w", client.portalURL, err)
	}
	if reply.status != http.StatusOK {
		return &StatusError{URL: redactQuery(client.portalURL + visitPath), Status: reply.status}
	}
	return nil
}

func (client *Client) getJSON(ctx context.Context, target string, value any) error {
	reply, err := client.getWithRetry(ctx, target)
	if err != nil {
		return err
	}
	if reply.status != http.StatusOK {
		return statusError(target, reply)
	}
	if err := json.Unmarshal(reply.body, value); err != nil {
		return &UnexpectedResponseError{URL: redactQuery(target), Status: reply.status, Err: err}
	}
	return nil
}

func (client *Client) getText(ctx context.Context, target string) (string, error) {
	reply, err := client.getWithRetry(ctx, target)
	if err != nil {
		return "", err
	}
	if reply.status != http.StatusOK {
		return "", statusError(target, reply)
	}
	return string(reply.body), nil
}

func (client *Client) getWithRetry(ctx context.Context, target string) (response, error) {
	var reply response
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var err error
		reply, err = client.transport.get(ctx, target)
		if err != nil {
			return response{}, err
		}
		if reply.status != http.StatusOK && isChallenged(reply.body) {
			return response{}, ErrChallenged
		}
		if !retryableStatuses[reply.status] {
			return reply, nil
		}
		client.logger.Warn("udemy is temporarily unavailable, retrying", "status", reply.status, "attempt", attempt, "url", redactQuery(target))
		if err := sleep(ctx, client.retryDelay); err != nil {
			return response{}, err
		}
	}
	return reply, nil
}

func statusError(target string, reply response) error {
	var problem struct {
		Detail string `json:"detail"`
	}
	_ = json.Unmarshal(reply.body, &problem)
	return &StatusError{URL: redactQuery(target), Status: reply.status, Detail: problem.Detail}
}

func isChallenged(body []byte) bool {
	for _, marker := range challengeMarkers {
		if bytes.Contains(body, marker) {
			return true
		}
	}
	return false
}

func sleep(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
