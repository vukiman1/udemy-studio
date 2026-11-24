package login

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const acceptedToken = "pasted-token-accepted" //nolint:gosec // reason: fake token for the fake portal

type memoryPasted struct {
	logins       []PastedLogin
	isUnreadable bool
}

func (store *memoryPasted) LoadPastedLogins(context.Context) ([]PastedLogin, error) {
	if store.isUnreadable {
		return nil, fmt.Errorf("%w: sealed with another key", ErrUnreadablePastedLogins)
	}
	return store.logins, nil
}

func (store *memoryPasted) SavePastedLogins(_ context.Context, logins []PastedLogin) error {
	store.logins, store.isUnreadable = logins, false
	return nil
}

// fakeChecker accepts only acceptedToken, as the portal would.
type fakeChecker struct{}

func (fakeChecker) UserName(_ context.Context, host string, token secret.Secret) (string, error) {
	if token.Reveal() != acceptedToken {
		return "", fmt.Errorf("%w by %s", ErrTokenRejected, host)
	}
	return "Learner", nil
}

func TestAPastedTokenIsUsedBeforeChromeAndWithoutIt(t *testing.T) {
	for _, selected := range []string{"Profile 4", ""} {
		service, _ := newService(selected)
		if _, err := service.SavePastedLogin(t.Context(), "https://acme.udemy.com/course/go/learn/", acceptedToken); err != nil {
			t.Fatalf("SavePastedLogin: %v", err)
		}
		if token, err := service.Token(t.Context(), "acme.udemy.com"); err != nil || token.Reveal() != acceptedToken {
			t.Errorf("with profile %q: Token = %q, %v; want the pasted token", selected, token.Reveal(), err)
		}
	}
}

func TestATokenThePortalRefusesIsNotSaved(t *testing.T) {
	service, _ := newService("")
	if _, err := service.SavePastedLogin(t.Context(), "acme.udemy.com", "copied-wrong"); !errors.Is(err, ErrTokenRejected) || !NeedsUserAction(err) {
		t.Errorf("SavePastedLogin = %v; want ErrTokenRejected, for the learner to fix", err)
	}
	if logins, _ := service.PastedLogins(t.Context()); len(logins) != 0 {
		t.Errorf("logins = %+v; want none saved", logins)
	}
}

func TestOnlyUdemyPortalsTakeAPastedToken(t *testing.T) {
	service, _ := newService("")
	for _, host := range []string{"evil.example.com", "acme.udemy.com.evil.com", ""} {
		if _, err := service.SavePastedLogin(t.Context(), host, acceptedToken); !errors.Is(err, ErrInvalidHost) {
			t.Errorf("SavePastedLogin(%q) = %v; want ErrInvalidHost", host, err)
		}
	}
}

func TestPastedTokensThatCannotBeOpenedCanBePastedAgain(t *testing.T) {
	service, _ := newService("")
	service.pasted = &memoryPasted{isUnreadable: true}
	if logins, err := service.PastedLogins(t.Context()); err != nil || len(logins) != 0 {
		t.Fatalf("PastedLogins = %+v, %v; want none", logins, err)
	}
	if _, err := service.SavePastedLogin(t.Context(), "acme.udemy.com", acceptedToken); err != nil {
		t.Errorf("SavePastedLogin over unreadable tokens: %v", err)
	}
}
