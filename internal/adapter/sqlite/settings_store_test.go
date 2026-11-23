package sqlite

import (
	"path/filepath"
	"testing"
)

func TestSelectedProfileIsRememberedAndReplaced(t *testing.T) {
	database, err := Open(t.Context(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	store := NewSettingsStore(database)

	if _, ok, err := store.SelectedProfile(t.Context()); err != nil || ok {
		t.Fatalf("fresh database: ok=%t err=%v, want nothing selected", ok, err)
	}
	for _, profileDir := range []string{"Profile 13", "Profile 4"} {
		if err := store.SaveSelectedProfile(t.Context(), profileDir); err != nil {
			t.Fatalf("save %s: %v", profileDir, err)
		}
	}
	if selected, ok, err := store.SelectedProfile(t.Context()); err != nil || !ok || selected != "Profile 4" {
		t.Errorf("selected = %q ok=%t err=%v, want Profile 4", selected, ok, err)
	}
}
