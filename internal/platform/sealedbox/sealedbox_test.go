package sealedbox

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	apiKey = "sk-ant-api-not-a-real-key" //nolint:gosec // reason: fake key proving it is not stored readable
	label  = "anthropic_api_key"
)

func TestSealedValueOpensOnlyWithItsKeyFileAndLabel(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "settings-encryption.key")
	box, err := LoadOrCreate(keyPath)
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	sealed, err := box.Seal([]byte(apiKey), label)
	if err != nil || strings.Contains(sealed, apiKey) {
		t.Fatalf("Seal = %q, %v; want ciphertext", sealed, err)
	}
	reloaded, err := LoadOrCreate(keyPath)
	if err != nil {
		t.Fatalf("reload key: %v", err)
	}
	if opened, err := reloaded.Open(sealed, label); err != nil || !bytes.Equal(opened, []byte(apiKey)) {
		t.Errorf("Open with the same key file = %q, %v", opened, err)
	}
	if _, err := reloaded.Open(sealed, "other_setting"); !errors.Is(err, ErrUnreadable) {
		t.Errorf("Open under another label: %v, want ErrUnreadable", err)
	}
	other, err := LoadOrCreate(filepath.Join(t.TempDir(), "other.key"))
	if err != nil {
		t.Fatalf("other key: %v", err)
	}
	if _, err := other.Open(sealed, label); !errors.Is(err, ErrUnreadable) {
		t.Errorf("Open with another key file: %v, want ErrUnreadable", err)
	}
	info, err := os.Stat(keyPath)
	if err != nil || info.Mode().Perm() != keyFileMode {
		t.Errorf("key file mode = %v, %v; want %v", info.Mode().Perm(), err, os.FileMode(keyFileMode))
	}
}
