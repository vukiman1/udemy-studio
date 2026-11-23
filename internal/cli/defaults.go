package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/vukiman1/udemy-studio/internal/app"
)

const (
	appDataDirName      = "udemy-studio"
	chromeDataDirSuffix = "Library/Application Support/Google/Chrome"
	defaultCaptionLang  = "en"
	defaultAria2c       = "aria2c"
	defaultYTDLP        = "yt-dlp"
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

// withDataDefaults fills the data and Chrome directories the flags left empty.
func withDataDefaults(dataDir, chromeDataDir string) (string, string, error) {
	var err error
	if dataDir == "" {
		if dataDir, err = defaultDataDir(); err != nil {
			return "", "", err
		}
	}
	if chromeDataDir == "" {
		if chromeDataDir, err = defaultChromeDataDir(); err != nil {
			return "", "", err
		}
	}
	return dataDir, chromeDataDir, nil
}

func addToolFlags(command *cobra.Command, tools *app.ToolPaths) {
	command.Flags().StringVar(&tools.Aria2c, "aria2c", defaultAria2c, "aria2c binary the Go downloader runs")
	command.Flags().StringVar(&tools.YTDLP, "yt-dlp", defaultYTDLP, "yt-dlp binary the Go downloader runs")
}
