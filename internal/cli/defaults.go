package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	appDataDirName      = "udemy-studio"
	chromeDataDirSuffix = "Library/Application Support/Google/Chrome"
	defaultCaptionLang  = "en"
	dataDirUsage        = "directory for the progress database, LAN access code and download logs (default: user config dir)"
	chromeDataDirUsage  = "Chrome user data dir (default: ~/" + chromeDataDirSuffix + ")"
)

func defaultDataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir for app data: %w", err)
	}
	return filepath.Join(configDir, appDataDirName), nil
}

func defaultChromeDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir for Chrome data: %w", err)
	}
	return filepath.Join(home, chromeDataDirSuffix), nil
}

func applyDataDefaults(dataDir, chromeDataDir *string) error {
	var err error
	if *dataDir == "" {
		if *dataDir, err = defaultDataDir(); err != nil {
			return err
		}
	}
	if *chromeDataDir == "" {
		if *chromeDataDir, err = defaultChromeDataDir(); err != nil {
			return err
		}
	}
	return nil
}
