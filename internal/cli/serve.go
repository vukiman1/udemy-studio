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
	libraryDirEnv        = "UDEMY_STUDIO_LIBRARY"
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
			var err error
			options.DataDir, options.Downloads.ChromeDataDir, err = withDataDefaults(options.DataDir, options.Downloads.ChromeDataDir)
			if err != nil {
				return err
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
	flags.StringVar(&options.DataDir, "data-dir", "", dataDirUsage)
	flags.BoolVar(&options.LANMode, "lan", false, "listen on the local network; other devices need the printed access code")
	addToolFlags(command, &options.Downloads.Tools)
	flags.StringVar(&options.Downloads.OutputDir, "download-dir", "", "where downloads are written (default: --library)")
	flags.StringVar(&options.Downloads.ChromeDataDir, "chrome-data-dir", "", chromeDataDirUsage)
	flags.StringVar(&options.Compression.FFmpeg, "ffmpeg", defaultFFmpeg, "ffmpeg binary that compresses videos to H.265")
	flags.StringVar(&options.Compression.FFprobe, "ffprobe", defaultFFprobe, "ffprobe binary that reads video codecs and lengths")
	flags.StringVar(&options.Downloads.CaptionLanguage, "caption-language", defaultCaptionLang, "caption language to download")
	return command
}

func resolvePaths(options *app.Options) error {
	paths := []*string{
		&options.LibraryDir,
		&options.DataDir,
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
