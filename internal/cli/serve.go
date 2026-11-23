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
)

func newServeCommand(version string) *cobra.Command {
	options := app.Options{Version: version}
	command := &cobra.Command{
		Use:   "serve",
		Short: "Run the web app over a downloads/ directory",
		RunE: func(command *cobra.Command, _ []string) error {
			if options.DatabasePath == "" {
				path, err := defaultDatabasePath()
				if err != nil {
					return err
				}
				options.DatabasePath = path
			}
			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			return app.Run(command.Context(), options, logger)
		},
	}
	flags := command.Flags()
	flags.StringVar(&options.ListenAddress, "addr", defaultListenAddress, "listen address")
	flags.StringVar(&options.LibraryDir, "library", os.Getenv(libraryDirEnv),
		"downloads/ directory produced by the downloader, read only (env "+libraryDirEnv+")")
	flags.StringVar(&options.DatabasePath, "db", "", "SQLite file for progress (default: user config dir)")
	flags.BoolVar(&options.LANMode, "lan", false, "listen on the local network; other devices need the printed access code")
	return command
}

func defaultDatabasePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir for the database: %w", err)
	}
	return filepath.Join(configDir, appDataDirName, databaseFileName), nil
}
