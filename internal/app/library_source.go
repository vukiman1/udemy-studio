package app

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"time"

	"github.com/vukiman1/udemy-studio/internal/cloudlibrary"
	"github.com/vukiman1/udemy-studio/internal/objectstore"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi"
)

const cloudRefreshInterval = 5 * time.Minute

type librarySource struct {
	files fs.FS
	links httpapi.MediaLinks
	cloud *cloudlibrary.Library
}

func openLibrarySource(ctx context.Context, role Role, media *os.Root, storage *objectstore.Service, logger *slog.Logger) librarySource {
	if role != RoleCloud {
		return librarySource{files: media.FS()}
	}
	cloud := cloudlibrary.New(storage, logger)
	if _, err := cloud.Refresh(ctx); err != nil {
		logger.Warn("cloud library not read at start; retrying every few minutes", "err", err)
	}
	return librarySource{files: cloud.FS(), links: cloud, cloud: cloud}
}

func (source librarySource) syncer(onChanged func(context.Context) error) func(context.Context) error {
	if source.cloud == nil {
		return nil
	}
	return func(ctx context.Context) error { return source.cloud.Sync(ctx, onChanged) }
}

func (source librarySource) workers(onChanged func(context.Context) error) []func(context.Context) error {
	if source.cloud == nil {
		return nil
	}
	return []func(context.Context) error{func(ctx context.Context) error {
		return source.cloud.Watch(ctx, cloudRefreshInterval, onChanged)
	}}
}
