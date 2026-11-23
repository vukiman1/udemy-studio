//go:build darwin

package chromecookie

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/vukiman1/udemy-studio/internal/login"
)

const keychainService = "Chrome Safe Storage"

type Keychain struct{}

func (Keychain) SafeStoragePassword(ctx context.Context) ([]byte, error) {
	password, err := exec.CommandContext(ctx, "security", "find-generic-password", "-w", "-s", keychainService).Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return nil, fmt.Errorf("%w: read item %q (allow access when macOS asks): %s (%w)", login.ErrNoKeychainAccess, keychainService, bytes.TrimSpace(exitErr.Stderr), err)
	}
	if err != nil {
		return nil, fmt.Errorf("run security to read keychain item %q: %w", keychainService, err)
	}
	return bytes.TrimSpace(password), nil
}
