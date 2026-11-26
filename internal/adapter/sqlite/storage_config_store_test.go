package sqlite

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite/gen"
	"github.com/vukiman1/udemy-studio/internal/objectstore"
	"github.com/vukiman1/udemy-studio/internal/platform/sealedbox"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const storedSecretAccessKey = "r2-secret-not-a-real-key-5432" //nolint:gosec // reason: fake key proving it is not stored readable

func TestStorageConfigIsStoredSealedAndReadsBack(t *testing.T) {
	database, err := Open(t.Context(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	box, err := sealedbox.LoadOrCreate(filepath.Join(t.TempDir(), "settings.key"))
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	store := NewStorageConfigStore(database, box)
	saved := objectstore.Connection{
		Location:    objectstore.Location{Endpoint: "https://abc123.r2.cloudflarestorage.com", Bucket: "studio-dev"},
		Credentials: objectstore.Credentials{AccessKeyID: "key-id", SecretAccessKey: secret.New(storedSecretAccessKey)},
	}
	if err := store.Save(t.Context(), saved); err != nil {
		t.Fatalf("Save: %v", err)
	}
	raw, err := gen.New(database).GetSetting(t.Context(), storageConfigSetting)
	if err != nil || strings.Contains(raw, storedSecretAccessKey) || strings.Contains(raw, "studio-dev") {
		t.Errorf("stored row = %q, %v; want it sealed", raw, err)
	}
	loaded, found, err := store.Load(t.Context())
	if err != nil || !found || loaded.Location != saved.Location || loaded.Credentials.SecretAccessKey.Reveal() != storedSecretAccessKey {
		t.Errorf("Load = %+v, %v, %v", loaded, found, err)
	}
	if err := store.Delete(t.Context()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, found, _ := store.Load(t.Context()); found {
		t.Error("config still found after Delete")
	}
}
