package sqlite

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite/gen"
	"github.com/vukiman1/udemy-studio/internal/login"
	"github.com/vukiman1/udemy-studio/internal/platform/sealedbox"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const pastedToken = "udemy-token-not-real-5432" //nolint:gosec // reason: fake token proving it is not stored readable

func TestPastedTokensAreStoredSealedAndReadBack(t *testing.T) {
	database, err := Open(t.Context(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	box, err := sealedbox.LoadOrCreate(filepath.Join(t.TempDir(), "settings.key"))
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	store := NewPastedLoginStore(database, box)
	savedAt := time.Date(2026, 11, 24, 9, 0, 0, 0, time.UTC)
	saved := []login.PastedLogin{{Host: "acme.udemy.com", Token: secret.New(pastedToken), UserName: "Learner", SavedAt: savedAt}}
	if err := store.SavePastedLogins(t.Context(), saved); err != nil {
		t.Fatalf("SavePastedLogins: %v", err)
	}
	raw, err := gen.New(database).GetSetting(t.Context(), pastedLoginsSetting)
	if err != nil || strings.Contains(raw, pastedToken) || strings.Contains(raw, "acme") {
		t.Errorf("stored row = %q, %v; want it sealed", raw, err)
	}
	loaded, err := store.LoadPastedLogins(t.Context())
	if err != nil || len(loaded) != 1 || loaded[0].Token.Reveal() != pastedToken || loaded[0].Host != "acme.udemy.com" || !loaded[0].SavedAt.Equal(savedAt) {
		t.Errorf("LoadPastedLogins = %+v, %v", loaded, err)
	}
}
