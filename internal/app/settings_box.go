package app

import (
	"path/filepath"

	"github.com/vukiman1/udemy-studio/internal/platform/sealedbox"
)

// The key that seals settings such as API keys and Udemy tokens lives beside the database, never inside it.
const settingsKeyFileName = "settings-encryption.key"

func openSettingsBox(dataDir string) (*sealedbox.Box, error) {
	return sealedbox.LoadOrCreate(filepath.Join(dataDir, settingsKeyFileName))
}
