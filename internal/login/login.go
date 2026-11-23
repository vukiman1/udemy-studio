package login

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

var (
	ErrNoProfileSelected = errors.New("no Chrome profile selected")
	ErrProfileNotFound   = errors.New("chrome profile not found")
	ErrNotLoggedIn       = errors.New("not logged in to Udemy")
	ErrTokenExpired      = errors.New("udemy login expired")
	ErrNoDiskAccess      = errors.New("no permission to read Chrome data")
	ErrNoKeychainAccess  = errors.New("no access to Chrome's key in the macOS keychain")

	userFixableErrors = []error{ErrNoProfileSelected, ErrProfileNotFound, ErrNotLoggedIn, ErrTokenExpired, ErrNoDiskAccess, ErrNoKeychainAccess}
)

// NeedsUserAction reports whether the learner fixes err in Chrome or System Settings rather than it being a fault of the app.
func NeedsUserAction(err error) bool {
	return slices.ContainsFunc(userFixableErrors, func(target error) bool { return errors.Is(err, target) })
}

type Profile struct {
	Dir        string
	Name       string
	UdemyHosts []string
}

type Token struct {
	Value     secret.Secret
	ExpiresAt time.Time
}

type CookieStore interface {
	Profiles(ctx context.Context) ([]Profile, error)
	Profile(ctx context.Context, profileDir string) (Profile, error)
	AccessToken(ctx context.Context, profileDir, host string) (Token, error)
}

type SettingsStore interface {
	SelectedProfile(ctx context.Context) (string, bool, error)
	SaveSelectedProfile(ctx context.Context, profileDir string) error
}

type Status struct {
	Profile Profile
	Hosts   []HostStatus
}

type HostStatus struct {
	Host      string
	ExpiresAt time.Time
	IsExpired bool
}

type Service struct {
	cookies  CookieStore
	settings SettingsStore
	now      func() time.Time
}

func NewService(cookies CookieStore, settings SettingsStore, now func() time.Time) *Service {
	return &Service{cookies: cookies, settings: settings, now: now}
}

func (service *Service) Profiles(ctx context.Context) ([]Profile, error) {
	return service.cookies.Profiles(ctx)
}

func (service *Service) SelectedProfileDir(ctx context.Context) (string, bool, error) {
	profileDir, ok, err := service.settings.SelectedProfile(ctx)
	if err != nil {
		return "", false, fmt.Errorf("load selected profile: %w", err)
	}
	return profileDir, ok, nil
}

func (service *Service) SelectProfile(ctx context.Context, profileDir string) error {
	if _, err := service.findProfile(ctx, profileDir); err != nil {
		return err
	}
	if err := service.settings.SaveSelectedProfile(ctx, profileDir); err != nil {
		return fmt.Errorf("save selected profile %q: %w", profileDir, err)
	}
	return nil
}

func (service *Service) Status(ctx context.Context) (Status, error) {
	profile, err := service.selectedProfile(ctx)
	if err != nil {
		return Status{}, err
	}
	status := Status{Profile: profile}
	for _, host := range profile.UdemyHosts {
		token, err := service.cookies.AccessToken(ctx, profile.Dir, host)
		if err != nil {
			return Status{}, fmt.Errorf("read login for %s in %q: %w", host, profile.Name, err)
		}
		status.Hosts = append(status.Hosts, HostStatus{Host: host, ExpiresAt: token.ExpiresAt, IsExpired: service.isExpired(token)})
	}
	return status, nil
}

func (service *Service) Token(ctx context.Context, host string) (secret.Secret, error) {
	profile, err := service.selectedProfile(ctx)
	if err != nil {
		return secret.Secret{}, err
	}
	token, err := service.cookies.AccessToken(ctx, profile.Dir, host)
	if errors.Is(err, ErrNotLoggedIn) {
		return secret.Secret{}, fmt.Errorf("%w: log in to %s in Chrome profile %q (%s)", ErrNotLoggedIn, host, profile.Name, profile.Dir)
	}
	if err != nil {
		return secret.Secret{}, fmt.Errorf("read login for %s in %q: %w", host, profile.Name, err)
	}
	if service.isExpired(token) {
		return secret.Secret{}, fmt.Errorf("%w: log in to %s again in Chrome profile %q (%s)", ErrTokenExpired, host, profile.Name, profile.Dir)
	}
	return token.Value, nil
}

func (service *Service) selectedProfile(ctx context.Context) (Profile, error) {
	profileDir, ok, err := service.settings.SelectedProfile(ctx)
	if err != nil {
		return Profile{}, fmt.Errorf("load selected profile: %w", err)
	}
	if !ok {
		return Profile{}, ErrNoProfileSelected
	}
	return service.findProfile(ctx, profileDir)
}

func (service *Service) findProfile(ctx context.Context, profileDir string) (Profile, error) {
	profile, err := service.cookies.Profile(ctx, profileDir)
	if err != nil {
		return Profile{}, fmt.Errorf("read Chrome profile %q: %w", profileDir, err)
	}
	return profile, nil
}

func (service *Service) isExpired(token Token) bool {
	return !token.ExpiresAt.IsZero() && !token.ExpiresAt.After(service.now())
}
