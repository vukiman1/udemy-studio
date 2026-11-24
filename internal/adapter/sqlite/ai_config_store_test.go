package sqlite

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite/gen"
	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/platform/sealedbox"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const storedKey = "sk-ant-api-not-a-real-key-9876" //nolint:gosec // reason: fake key proving it is not stored readable

func TestAIConfigIsStoredSealedAndReadsBack(t *testing.T) {
	database, err := Open(t.Context(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	box, err := sealedbox.LoadOrCreate(filepath.Join(t.TempDir(), "settings.key"))
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	store := NewAIConfigStore(database, box)
	saved := llm.Config{
		Active:    llm.ProviderClaude,
		Providers: map[llm.Provider]llm.ProviderSettings{llm.ProviderClaude: {APIKey: secret.New(storedKey), Models: []string{"claude-opus-5-5"}}},
		IsEnabled: true,
	}
	if err := store.Save(t.Context(), saved); err != nil {
		t.Fatalf("Save: %v", err)
	}
	raw, err := gen.New(database).GetSetting(t.Context(), aiConfigSetting)
	if err != nil || strings.Contains(raw, storedKey) || strings.Contains(raw, "claude") {
		t.Errorf("stored row = %q, %v; want it sealed", raw, err)
	}
	loaded, found, err := store.Load(t.Context())
	if err != nil || !found || loaded.Providers[llm.ProviderClaude].APIKey.Reveal() != storedKey || loaded.Active != saved.Active || !loaded.IsEnabled {
		t.Errorf("Load = %+v, %v, %v", loaded, found, err)
	}
	if err := store.Delete(t.Context()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, found, _ := store.Load(t.Context()); found {
		t.Error("config still found after Delete")
	}
}
