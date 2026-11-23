package login

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

var now = time.Date(2026, 11, 23, 8, 0, 0, 0, time.UTC)

type fakeCookies struct {
	profiles []Profile
	tokens   map[string]Token
}

func (fake fakeCookies) Profiles(context.Context) ([]Profile, error) { return fake.profiles, nil }

func (fake fakeCookies) Profile(_ context.Context, profileDir string) (Profile, error) {
	for _, profile := range fake.profiles {
		if profile.Dir == profileDir {
			return profile, nil
		}
	}
	return Profile{}, ErrProfileNotFound
}

func (fake fakeCookies) AccessToken(_ context.Context, profileDir, host string) (Token, error) {
	token, ok := fake.tokens[profileDir+"|"+host]
	if !ok {
		return Token{}, ErrNotLoggedIn
	}
	return token, nil
}

type fakeSettings struct{ selected string }

func (fake *fakeSettings) SelectedProfile(context.Context) (string, bool, error) {
	return fake.selected, fake.selected != "", nil
}

func (fake *fakeSettings) SaveSelectedProfile(_ context.Context, profileDir string) error {
	fake.selected = profileDir
	return nil
}

func newService(selected string) (*Service, *fakeSettings) {
	cookies := fakeCookies{
		profiles: []Profile{
			{Dir: "Profile 4", Name: "work", UdemyHosts: []string{"acme.udemy.com"}},
			{Dir: "Profile 13", Name: "work", UdemyHosts: nil},
		},
		tokens: map[string]Token{
			"Profile 4|acme.udemy.com": {Value: secret.New("fresh"), ExpiresAt: now.Add(time.Hour)},
			"Profile 4|old.udemy.com":  {Value: secret.New("stale"), ExpiresAt: now.Add(-time.Minute)},
		},
	}
	settings := &fakeSettings{selected: selected}
	return NewService(cookies, settings, func() time.Time { return now }), settings
}

func TestTokenComesFromTheSelectedProfile(t *testing.T) {
	service, _ := newService("Profile 4")
	token, err := service.Token(t.Context(), "acme.udemy.com")
	if err != nil || token.Reveal() != "fresh" {
		t.Fatalf("Token = %v, %v", token.Reveal(), err)
	}
}

func TestTokenErrorsNameTheProfileToFix(t *testing.T) {
	cases := []struct {
		name     string
		selected string
		host     string
		want     error
	}{
		{name: "nothing selected", selected: "", host: "acme.udemy.com", want: ErrNoProfileSelected},
		{name: "selected profile disappeared", selected: "Profile 99", host: "acme.udemy.com", want: ErrProfileNotFound},
		{name: "not logged in to that portal", selected: "Profile 13", host: "acme.udemy.com", want: ErrNotLoggedIn},
		{name: "login expired", selected: "Profile 4", host: "old.udemy.com", want: ErrTokenExpired},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			service, _ := newService(testCase.selected)
			_, err := service.Token(t.Context(), testCase.host)
			if !errors.Is(err, testCase.want) {
				t.Fatalf("error = %v, want %v", err, testCase.want)
			}
			if testCase.selected != "" && !errors.Is(testCase.want, ErrProfileNotFound) && !strings.Contains(err.Error(), testCase.selected) {
				t.Errorf("error %q does not name profile %q", err, testCase.selected)
			}
		})
	}
}

func TestSelectProfileOnlyAcceptsExistingProfiles(t *testing.T) {
	service, settings := newService("")
	if err := service.SelectProfile(t.Context(), "Profile 99"); !errors.Is(err, ErrProfileNotFound) {
		t.Errorf("select unknown = %v, want ErrProfileNotFound", err)
	}
	if err := service.SelectProfile(t.Context(), "Profile 13"); err != nil || settings.selected != "Profile 13" {
		t.Errorf("select known = %v, saved %q", err, settings.selected)
	}
}

func TestStatusReportsEachLoggedInPortal(t *testing.T) {
	service, _ := newService("Profile 4")
	status, err := service.Status(t.Context())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.Profile.Dir != "Profile 4" || len(status.Hosts) != 1 || status.Hosts[0].IsExpired || !status.Hosts[0].ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Errorf("status = %+v", status)
	}
}
