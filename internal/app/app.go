package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite"
	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/notes"
	"github.com/vukiman1/udemy-studio/internal/progress"
	"github.com/vukiman1/udemy-studio/internal/search"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi"
	"github.com/vukiman1/udemy-studio/web"
)

const (
	readHeaderTimeout   = 10 * time.Second
	shutdownGracePeriod = 3 * time.Second
	dataDirMode         = 0o750
)

var ErrMissingLibraryDir = errors.New("library directory is required")

type Options struct {
	ListenAddress string
	Version       string
	LibraryDir    string
	DataDir       string
	LANMode       bool
	// PublicHosts are hostnames a tunnel or proxy serves the app under, trusted besides localhost and IPs.
	PublicHosts []string
	Downloads   DownloadOptions
	Compression CompressionOptions
}

func Run(ctx context.Context, options Options, logger *slog.Logger) error {
	if options.LibraryDir == "" {
		return ErrMissingLibraryDir
	}
	media, err := os.OpenRoot(options.LibraryDir)
	if err != nil {
		return fmt.Errorf("open library %s: %w", options.LibraryDir, err)
	}
	defer media.Close()
	catalog, err := library.NewCatalog(media.FS())
	if err != nil {
		return fmt.Errorf("scan library %s: %w", options.LibraryDir, err)
	}
	logger.Info("library scanned", "dir", options.LibraryDir, "courses", len(catalog.Current().Courses))

	database, err := openDatabase(ctx, options.DataDir)
	if err != nil {
		return err
	}
	defer database.Close()
	progressService := progress.NewService(sqlite.NewProgressStore(database), time.Now)
	notesService := notes.NewService(sqlite.NewNotesStore(database), time.Now)
	if err := adoptLectureKeys(ctx, catalog, progressService, notesService); err != nil {
		return err
	}

	frontend, err := web.Assets()
	if errors.Is(err, web.ErrFrontendNotBuilt) {
		logger.Warn("serving API only", "reason", err)
		frontend = nil
	} else if err != nil {
		return fmt.Errorf("load embedded frontend: %w", err)
	}

	compression := newCompression(options.Compression, options.LibraryDir, media, database, catalog, logger)
	refresher := libraryRefresher{
		catalog:     catalog,
		media:       media.FS(),
		progress:    progressService,
		notes:       notesService,
		search:      search.NewService(sqlite.NewCueIndex(database), logger),
		compression: compression,
		logger:      logger,
	}
	box, err := openSettingsBox(options.DataDir)
	if err != nil {
		return err
	}
	downloads, err := newDownloads(ctx, options.Downloads, options.DataDir, database, box, refresher.afterLibraryChanged, logger)
	if err != nil {
		return err
	}
	ai := newAISettings(box, database, logger)
	storage := newStorageSettings(box, database, logger)
	translation := newTranslationQueue(media, refresher, ai, logger)
	handler := httpapi.NewHandler(httpapi.Deps{
		Version:     options.Version,
		Frontend:    frontend,
		Catalog:     catalog,
		Media:       media.FS(),
		Progress:    progressService,
		Search:      refresher.search,
		Notes:       notesService,
		Compression: compression,
		Translation: translation,
		AI:          ai,
		Storage:     storage,
		Login:       downloads.login,
		Jobs:        downloads.jobs,
		Logger:      logger,
		PublicHosts: options.PublicHosts,
	})
	address, accessCode := options.ListenAddress, ""
	if options.LANMode {
		address, handler, accessCode, err = enableLANMode(address, handler, options.DataDir)
		if err != nil {
			return err
		}
	}
	listener, err := new(net.ListenConfig).Listen(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}
	if options.LANMode {
		if err := announceLANMode(os.Stdout, listener.Addr(), accessCode); err != nil {
			_ = listener.Close()
			return err
		}
	}
	logger.Info("http server listening", "addr", listener.Addr().String(), "version", options.Version)
	workers := append(downloads.workers(), refresher.indexSubtitles, translation.Run)
	if compression != nil {
		workers = append(workers, compression.Run, compression.Survey)
	}
	return serve(ctx, listener, handler, logger, shutdownGracePeriod, workers...)
}

func serve(
	ctx context.Context,
	listener net.Listener,
	handler http.Handler,
	logger *slog.Logger,
	gracePeriod time.Duration,
	workers ...func(context.Context) error,
) error {
	server := &http.Server{Handler: handler, ReadHeaderTimeout: readHeaderTimeout}

	group, groupCtx := errgroup.WithContext(ctx)
	for _, worker := range workers {
		group.Go(func() error { return worker(groupCtx) })
	}
	group.Go(func() error {
		if err := server.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve http on %s: %w", listener.Addr(), err)
		}
		return nil
	})
	group.Go(func() error {
		<-groupCtx.Done()
		return shutdown(context.WithoutCancel(ctx), server, logger, gracePeriod)
	})
	return group.Wait()
}

// Video streams stay open for as long as the browser buffers, so after the grace period they are cut rather than awaited.
func shutdown(ctx context.Context, server *http.Server, logger *slog.Logger, gracePeriod time.Duration) error {
	logger.Info("http server shutting down", "grace_period", gracePeriod)
	graceCtx, cancel := context.WithTimeout(ctx, gracePeriod)
	defer cancel()
	err := server.Shutdown(graceCtx)
	if !errors.Is(err, context.DeadlineExceeded) {
		if err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		return nil
	}
	logger.Info("closing connections still streaming after the grace period")
	if err := server.Close(); err != nil {
		return fmt.Errorf("force close http server: %w", err)
	}
	return nil
}
