package transcode

import (
	"context"
	"io/fs"
	"sync"
	"time"
)

// cachedMedia is a video as ffprobe last read it, valid while the file keeps that size and modification time.
type cachedMedia struct {
	size     int64
	modTime  time.Time
	media    Media
	probeErr error
}

type mediaCache struct {
	mutex   sync.Mutex
	entries map[string]cachedMedia
}

func (cache *mediaCache) lookup(path string, info fs.FileInfo) (cachedMedia, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	entry, ok := cache.entries[path]
	if !ok || entry.size != info.Size() || !entry.modTime.Equal(info.ModTime()) {
		return cachedMedia{}, false
	}
	return entry, true
}

func (cache *mediaCache) store(path string, info fs.FileInfo, media Media, probeErr error) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.entries[path] = cachedMedia{size: info.Size(), modTime: info.ModTime(), media: media, probeErr: probeErr}
}

// probe reads a video, asking ffprobe only when the file changed since it was last read successfully; a failure may
// have been passing, so it is asked again, while the library page still sees it as unreadable rather than unchecked.
func (service *Service) probe(ctx context.Context, path string) (Media, error) {
	info, err := service.deps.Files.Stat(path)
	if err != nil {
		return Media{}, err
	}
	if cached, ok := service.cache.lookup(path, info); ok && cached.probeErr == nil {
		return cached.media, nil
	}
	media, err := service.deps.Prober.Probe(ctx, path)
	if ctx.Err() != nil {
		return Media{}, ctx.Err()
	}
	service.cache.store(path, info, media, err)
	return media, err
}

// Survey reads every video of the library into the cache in the background, again whenever Resurvey asks,
// so the library page can tell what is left to compress without waiting for ffprobe.
func (service *Service) Survey(ctx context.Context) error {
	for {
		service.surveyLibrary(ctx)
		select {
		case <-ctx.Done():
			return nil
		case <-service.resurvey:
		}
	}
}

func (service *Service) surveyLibrary(ctx context.Context) {
	for _, course := range service.deps.Catalog.Current().Courses {
		for _, chapter := range course.Chapters {
			for _, lecture := range chapter.Lectures {
				select {
				case <-ctx.Done():
					return
				default:
				}
				if lecture.VideoPath != "" {
					service.surveyVideo(ctx, lecture.VideoPath)
				}
			}
		}
	}
}

// surveyVideo caches one video; one it cannot read is cached as unreadable, so it is not read again until it changes.
func (service *Service) surveyVideo(ctx context.Context, path string) {
	if _, err := service.probe(ctx, path); err != nil && ctx.Err() == nil {
		service.deps.Logger.Debug("survey could not read a video", "path", path, "err", err)
	}
}

// Resurvey asks the survey to read the library again, for files a download added or changed.
func (service *Service) Resurvey() {
	select {
	case service.resurvey <- struct{}{}:
	default:
	}
}
