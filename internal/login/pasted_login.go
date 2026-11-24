package login

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const udemyHostSuffix = ".udemy.com"

var (
	ErrInvalidHost   = errors.New("not a Udemy portal such as acme.udemy.com")
	ErrEmptyToken    = errors.New("paste the value of the access_token cookie")
	ErrTokenRejected = errors.New("udemy did not accept the pasted token")
	// ErrPortalUnreachable means the token could not be checked at all, which is no verdict on the token.
	ErrPortalUnreachable = errors.New("could not ask Udemy about the token")
	// ErrUnreadablePastedLogins means the saved tokens cannot be opened, for example after the sealing key was lost.
	ErrUnreadablePastedLogins = errors.New("the pasted Udemy tokens cannot be read")
)

// PastedLogin is an access_token copied by hand from a browser, for where Chrome cannot be read, such as in Docker.
type PastedLogin struct {
	Host     string
	Token    secret.Secret
	UserName string
	SavedAt  time.Time
}

type PastedLoginStore interface {
	LoadPastedLogins(ctx context.Context) ([]PastedLogin, error)
	SavePastedLogins(ctx context.Context, logins []PastedLogin) error
}

// TokenChecker asks the portal whose token it is; a token the portal refuses wraps ErrTokenRejected.
type TokenChecker interface {
	UserName(ctx context.Context, host string, token secret.Secret) (string, error)
}

func (service *Service) PastedLogins(ctx context.Context) ([]PastedLogin, error) {
	return service.loadPastedLogins(ctx)
}

// SavePastedLogin keeps a token for the portal only once the portal has accepted it, replacing the portal's earlier one.
func (service *Service) SavePastedLogin(ctx context.Context, hostOrURL, token string) (PastedLogin, error) {
	host, err := normalizeHost(hostOrURL)
	if err != nil {
		return PastedLogin{}, err
	}
	value := strings.TrimSpace(token)
	if value == "" {
		return PastedLogin{}, ErrEmptyToken
	}
	userName, err := service.checker.UserName(ctx, host, secret.New(value))
	if errors.Is(err, ErrTokenRejected) {
		return PastedLogin{}, fmt.Errorf("check the token for %s: %w", host, err)
	}
	if err != nil {
		return PastedLogin{}, fmt.Errorf("%w %s: %w", ErrPortalUnreachable, host, err)
	}
	logins, err := service.loadPastedLogins(ctx)
	if err != nil {
		return PastedLogin{}, err
	}
	saved := PastedLogin{Host: host, Token: secret.New(value), UserName: userName, SavedAt: service.now()}
	logins = append(withoutHost(logins, host), saved)
	if err := service.pasted.SavePastedLogins(ctx, logins); err != nil {
		return PastedLogin{}, fmt.Errorf("save the token for %s: %w", host, err)
	}
	return saved, nil
}

func (service *Service) ForgetPastedLogin(ctx context.Context, host string) error {
	logins, err := service.loadPastedLogins(ctx)
	if err != nil {
		return err
	}
	if err := service.pasted.SavePastedLogins(ctx, withoutHost(logins, host)); err != nil {
		return fmt.Errorf("forget the token for %s: %w", host, err)
	}
	return nil
}

func (service *Service) pastedToken(ctx context.Context, host string) (secret.Secret, bool, error) {
	logins, err := service.loadPastedLogins(ctx)
	if err != nil {
		return secret.Secret{}, false, err
	}
	position := slices.IndexFunc(logins, func(login PastedLogin) bool { return login.Host == host })
	if position < 0 {
		return secret.Secret{}, false, nil
	}
	return logins[position].Token, true, nil
}

// loadPastedLogins starts over from none when the saved tokens can no longer be opened, so pasting again fixes it.
func (service *Service) loadPastedLogins(ctx context.Context) ([]PastedLogin, error) {
	logins, err := service.pasted.LoadPastedLogins(ctx)
	if errors.Is(err, ErrUnreadablePastedLogins) {
		service.logger.Warn("starting the pasted Udemy tokens over: the saved ones cannot be read", "err", err)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load pasted Udemy tokens: %w", err)
	}
	return logins, nil
}

// normalizeHost accepts the portal's host or any URL on it, as copied from the address bar.
func normalizeHost(hostOrURL string) (string, error) {
	text := strings.ToLower(strings.TrimSpace(hostOrURL))
	if strings.Contains(text, "://") {
		parsed, err := url.Parse(text)
		if err != nil {
			return "", fmt.Errorf("%w: %q", ErrInvalidHost, hostOrURL)
		}
		text = parsed.Hostname()
	}
	if !strings.HasSuffix(text, udemyHostSuffix) || strings.ContainsAny(text, "/:@ ") {
		return "", fmt.Errorf("%w: %q", ErrInvalidHost, hostOrURL)
	}
	return text, nil
}

func withoutHost(logins []PastedLogin, host string) []PastedLogin {
	return slices.DeleteFunc(slices.Clone(logins), func(login PastedLogin) bool { return login.Host == host })
}
