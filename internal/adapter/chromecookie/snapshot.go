package chromecookie

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	cookiesFile      = "Cookies"
	snapshotFileMode = 0o600
)

var sqliteCompanionSuffixes = []string{"", "-wal", "-shm"}

// Chrome keeps Cookies locked and partly in its WAL while running, so queries run on a private copy.
func snapshotCookies(profilePath string) (string, func(), error) {
	snapshotDir, err := os.MkdirTemp("", "udemy-studio-cookies-*")
	if err != nil {
		return "", nil, fmt.Errorf("create cookie snapshot dir: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(snapshotDir) }
	for _, suffix := range sqliteCompanionSuffixes {
		err := copyFile(filepath.Join(profilePath, cookiesFile+suffix), filepath.Join(snapshotDir, cookiesFile+suffix))
		if suffix != "" && errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			cleanup()
			return "", nil, wrapAccessError(err)
		}
	}
	return filepath.Join(snapshotDir, cookiesFile), cleanup, nil
}

func copyFile(source, destination string) error {
	input, err := os.Open(source) //nolint:gosec // reason: source is a Chrome profile path validated against Local State
	if err != nil {
		return fmt.Errorf("open %s: %w", source, err)
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, snapshotFileMode) //nolint:gosec // reason: destination is inside our own temp dir
	if err != nil {
		return fmt.Errorf("create %s: %w", destination, err)
	}
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		return fmt.Errorf("copy %s: %w", source, err)
	}
	return output.Close()
}
