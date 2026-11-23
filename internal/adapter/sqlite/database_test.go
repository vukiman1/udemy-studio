package sqlite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenUsesPathsWithURIReservedCharactersVerbatim(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "notes #1 ?x 50%")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, "app.db")
	database, err := Open(t.Context(), path)
	if err != nil {
		t.Fatalf("Open(%q): %v", path, err)
	}
	_ = database.Close()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("database not created at the exact path: %v", err)
	}
}

func TestOpenIsIdempotentOnAnExistingDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")
	for attempt := 1; attempt <= 2; attempt++ {
		database, err := Open(t.Context(), path)
		if err != nil {
			t.Fatalf("Open attempt %d: %v", attempt, err)
		}
		_ = database.Close()
	}
}
