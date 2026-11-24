package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vukiman1/udemy-studio/internal/app"
)

// ErrInvalidPublicHost is a --public-host given as a URL or with a port instead of a bare hostname.
var ErrInvalidPublicHost = errors.New("--public-host takes a hostname such as studio.example.com, without scheme, port or path")

const (
	defaultListenAddress = "127.0.0.1:8765"
	libraryDirEnv        = "UDEMY_STUDIO_LIBRARY"
	publicHostEnv        = "UDEMY_STUDIO_PUBLIC_HOST"
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
			if options.PublicHosts, err = publicHosts(options.PublicHosts); err != nil {
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
	flags.StringSliceVar(&options.PublicHosts, "public-host", splitList(os.Getenv(publicHostEnv)),
		"hostname a tunnel or proxy serves the app under, such as studio.example.com; put a login such as Cloudflare Access in front (env "+publicHostEnv+")")
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

// publicHosts keeps bare hostnames only, lowercased as the host guard compares them.
func publicHosts(values []string) ([]string, error) {
	hosts := make([]string, 0, len(values))
	for _, value := range values {
		host := strings.ToLower(strings.TrimSpace(value))
		if host == "" {
			continue
		}
		if strings.ContainsAny(host, "/:@ ") {
			return nil, fmt.Errorf("%w: %q", ErrInvalidPublicHost, value)
		}
		hosts = append(hosts, host)
	}
	return hosts, nil
}

func splitList(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}
