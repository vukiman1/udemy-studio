//go:build darwin

package chromecookie

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

const keychainService = "Chrome Safe Storage"

type Keychain struct{}

func (Keychain) SafeStoragePassword(ctx context.Context) ([]byte, error) {
	password, err := exec.CommandContext(ctx, "security", "find-generic-password", "-w", "-s", keychainService).Output()
	if err != nil {
		return nil, fmt.Errorf("read keychain item %q (allow access when macOS asks): %w", keychainService, err)
	}
	return bytes.TrimSpace(password), nil
}
