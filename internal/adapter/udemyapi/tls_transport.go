package udemyapi

import (
	"context"
	"fmt"
	"io"
	"net/url"

	http "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	requestTimeoutSeconds = 120
	accessTokenCookie     = "access_token"
	maxResponseBytes      = 64 << 20
)

var errResponseTooLarge = fmt.Errorf("response is larger than %d bytes", maxResponseBytes)

// Udemy's Android app identity; the API expects it alongside the login cookie.
var appHeaders = map[string]string{ //nolint:gosec // reason: public client id/secret shipped in Udemy's Android app and the upstream downloader, not a user credential
	"accept-encoding":        "gzip",
	"x-mobile-visit-enabled": "true",
	"x-udemy-client-secret":  "f2lgDUDxjFiOlVHUpwQNFUfCQPyMO0tJQMaud53PF01UKueW8enYjeEYoyVeP0bb2XVEDkJ5GLJaVTfM5QgMVz6yyXyydZdA5QhzgvG9UmCPUYaCrIVf7VpmiilfbLJc",
	"x-udemy-client-id":      "TH96Ov3Ebo3OtgoSH5mOYzYolcowM3ycedWQDDce",
	"accept-language":        "en_US",
	"x-version-name":         "9.51.2",
	"x-client-name":          "Udemy-Android",
	"user-agent":             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/154.0.0.0 Safari/537.36",
}

type response struct {
	status int
	body   []byte
}

type transport interface {
	get(ctx context.Context, target string) (response, error)
}

// tlsTransport speaks with Chrome's TLS and HTTP/2 fingerprint so Cloudflare lets the requests through.
type tlsTransport struct {
	client tlsclient.HttpClient
}

func newTLSTransport(portalURL string, token secret.Secret) (tlsTransport, error) {
	client, err := tlsclient.NewHttpClient(tlsclient.NewNoopLogger(),
		tlsclient.WithTimeoutSeconds(requestTimeoutSeconds),
		tlsclient.WithClientProfile(profiles.Chrome_154),
		tlsclient.WithCookieJar(tlsclient.NewCookieJar()),
	)
	if err != nil {
		return tlsTransport{}, fmt.Errorf("create tls client: %w", err)
	}
	portal, err := url.Parse(portalURL)
	if err != nil {
		return tlsTransport{}, fmt.Errorf("parse portal url %q: %w", portalURL, err)
	}
	client.SetCookies(portal, []*http.Cookie{{Name: accessTokenCookie, Value: token.Reveal(), Path: "/"}})
	return tlsTransport{client: client}, nil
}

func (transport tlsTransport) get(ctx context.Context, target string) (response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return response{}, fmt.Errorf("build request for %s: %w", redactQuery(target), err)
	}
	for name, value := range appHeaders {
		request.Header.Set(name, value)
	}
	reply, err := transport.client.Do(request)
	if err != nil {
		return response{}, fmt.Errorf("get %s: %w", redactQuery(target), err)
	}
	defer reply.Body.Close()
	decoded := http.DecompressBody(reply)
	defer decoded.Close()
	body, err := io.ReadAll(io.LimitReader(decoded, maxResponseBytes+1))
	if err != nil {
		return response{}, fmt.Errorf("read %s: %w", redactQuery(target), err)
	}
	if len(body) > maxResponseBytes {
		return response{}, fmt.Errorf("read %s: %w", redactQuery(target), errResponseTooLarge)
	}
	return response{status: reply.StatusCode, body: body}, nil
}

func redactQuery(target string) string {
	parsed, err := url.Parse(target)
	if err != nil {
		return "<unparsable url>"
	}
	parsed.RawQuery = ""
	return parsed.String()
}
