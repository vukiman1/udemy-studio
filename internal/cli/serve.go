package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/vukiman1/udemy-studio/internal/app"
)

const (
	defaultListenAddress = "127.0.0.1:8765"
	appDataDirName       = "udemy-studio"
	databaseFileName     = "app.db"
	libraryDirEnv        = "UDEMY_STUDIO_LIBRARY"
	pythonEnv            = "UDEMY_STUDIO_PYTHON"
	pythonRepoEnv        = "UDEMY_STUDIO_PYTHON_REPO"
	chromeDataDirSuffix  = "Library/Application Support/Google/Chrome"
	defaultCaptionLang   = "en"
)

func newServeCommand(version string) *cobra.Command {
	options := app.Options{Version: version}
	command := &cobra.Command{
		Use:   "serve",
		Short: "Run the web app over a downloads/ directory",
		RunE: func(command *cobra.Command, _ []string) error {
			if options.Downloads.OutputDir == "" {
				options.Downloads.OutputDir = options.LibraryDir
			}
			if options.Downloads.ChromeDataDir == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("resolve home dir for Chrome data: %w", err)
				}
				options.Downloads.ChromeDataDir = filepath.Join(home, chromeDataDirSuffix)
			}
			if options.DatabasePath == "" {
				path, err := defaultDatabasePath()
				if err != nil {
					return err
				}
				options.DatabasePath = path
			}
			if err := resolvePaths(&options); err != nil {
				return err
			}
			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			return app.Run(command.Context(), options, logger)
		},
	}
	flags := command.Flags()
	flags.StringVar(&options.ListenAddress, "addr", defaultListenAddress, "listen address")
	flags.StringVar(&options.LibraryDir, "library", os.Getenv(libraryDirEnv),
		"downloads/ directory the player reads and downloads go to by default (env "+libraryDirEnv+")")
	flags.StringVar(&options.DatabasePath, "db", "", "SQLite file for progress (default: user config dir)")
	flags.BoolVar(&options.LANMode, "lan", false, "listen on the local network; other devices need the printed access code")
	flags.StringVar(&options.Downloads.Python, "python", os.Getenv(pythonEnv),
		"python of the udemy-downloader venv; enables downloads (env "+pythonEnv+")")
	flags.StringVar(&options.Downloads.PythonRepo, "python-repo", os.Getenv(pythonRepoEnv),
		"udemy-downloader checkout the downloader runs in (env "+pythonRepoEnv+")")
	flags.StringVar(&options.Downloads.OutputDir, "download-dir", "", "where downloads are written (default: --library)")
	flags.StringVar(&options.Downloads.ChromeDataDir, "chrome-data-dir", "", "Chrome user data dir (default: ~/"+chromeDataDirSuffix+")")
	flags.StringVar(&options.Downloads.CaptionLanguage, "caption-language", defaultCaptionLang, "caption language to download")
	return command
}

func defaultDatabasePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir for the database: %w", err)
	}
	return filepath.Join(configDir, appDataDirName, databaseFileName), nil
}

// The downloader runs inside its own checkout, so a relative path would point somewhere else for it.
func resolvePaths(options *app.Options) error {
	paths := []*string{
		&options.LibraryDir,
		&options.DatabasePath,
		&options.Downloads.Python,
		&options.Downloads.PythonRepo,
		&options.Downloads.OutputDir,
		&options.Downloads.ChromeDataDir,
	}
	for _, path := range paths {
		if *path == "" {
			continue
		}
		absolute, err := filepath.Abs(*path)
		if err != nil {
			return fmt.Errorf("resolve path %q: %w", *path, err)
		}
		*path = absolute
	}
	return nil
}
