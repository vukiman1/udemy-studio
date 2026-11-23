package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	appDataDirName      = "udemy-studio"
	databaseFileName    = "app.db"
	chromeDataDirSuffix = "Library/Application Support/Google/Chrome"
	defaultCaptionLang  = "en"
)

func applyDataDefaults(databasePath, chromeDataDir *string) error {
	if *chromeDataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolve home dir for Chrome data: %w", err)
		}
		*chromeDataDir = filepath.Join(home, chromeDataDirSuffix)
	}
	if *databasePath == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("resolve user config dir for the database: %w", err)
		}
		*databasePath = filepath.Join(configDir, appDataDirName, databaseFileName)
	}
	return nil
}
