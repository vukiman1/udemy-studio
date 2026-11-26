package cloudlibrary

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"sync"
	"time"

	"github.com/vukiman1/udemy-studio/internal/objectstore"
	"github.com/vukiman1/udemy-studio/internal/publish"
)

const LinkLifetime = 6 * time.Hour

var ErrNoBucket = errors.New("no cloud storage is configured yet")

type BucketSource interface {
	Bucket(ctx context.Context) (objectstore.Bucket, error)
}

type Library struct {
	buckets   BucketSource
	logger    *slog.Logger
	mutex     sync.RWMutex
	snapshot  snapshot
	syncMutex sync.Mutex
}

func New(buckets BucketSource, logger *slog.Logger) *Library {
	return &Library{buckets: buckets, logger: logger, snapshot: snapshot{tree: emptyTree()}}
}

func (library *Library) FS() fs.FS {
	return libraryFS{library: library}
}

func (library *Library) Refresh(ctx context.Context) (bool, error) {
	bucket, err := library.buckets.Bucket(ctx)
	if errors.Is(err, objectstore.ErrNotConfigured) {
		return library.replace(snapshot{tree: emptyTree()}), nil
	}
	if err != nil {
		return false, err
	}
	index, found, err := publish.ReadIndex(ctx, bucket)
	if err != nil {
		return false, err
	}
	next := snapshot{tree: emptyTree(), bucket: bucket}
	if found {
		next.tree = treeOf(index)
	}
	return library.replace(next), nil
}

func (library *Library) Watch(ctx context.Context, every time.Duration, onChanged func(context.Context) error) error {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := library.Sync(ctx, onChanged); err != nil {
				library.logger.Warn("cloud library not refreshed; keeping the last one", "err", err)
			}
		}
	}
}

func (library *Library) Sync(ctx context.Context, onChanged func(context.Context) error) error {
	library.syncMutex.Lock()
	defer library.syncMutex.Unlock()
	changed, err := library.Refresh(ctx)
	if err != nil || !changed {
		return err
	}
	library.logger.Info("cloud library changed", "files", len(library.current().tree.nodes))
	if err := onChanged(ctx); err != nil {
		return fmt.Errorf("follow the changed cloud library: %w", err)
	}
	return nil
}

func (library *Library) Link(ctx context.Context, filePath, contentDisposition string) (string, error) {
	current := library.current()
	if current.bucket == nil {
		return "", ErrNoBucket
	}
	link, err := current.bucket.Link(ctx, objectstore.LinkRequest{
		Key:                publish.CoursesPrefix + filePath,
		ExpiresIn:          LinkLifetime,
		ContentDisposition: contentDisposition,
	})
	if err != nil {
		return "", fmt.Errorf("link to %s: %w", filePath, err)
	}
	return link, nil
}

func (library *Library) current() snapshot {
	library.mutex.RLock()
	defer library.mutex.RUnlock()
	return library.snapshot
}

func (library *Library) replace(next snapshot) bool {
	library.mutex.Lock()
	defer library.mutex.Unlock()
	changed := !next.tree.modifiedAt.Equal(library.snapshot.tree.modifiedAt) || len(next.tree.nodes) != len(library.snapshot.tree.nodes)
	library.snapshot = next
	return changed
}
