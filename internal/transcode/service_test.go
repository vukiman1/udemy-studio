package transcode

import (
	"context"
	"io/fs"
	"log/slog"
	"maps"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
)

const (
	shrinks     = "course/01 - Intro/001 Shrinks.mp4"
	drifts      = "course/01 - Intro/002 Drifts.mp4"
	alreadyH265 = "course/01 - Intro/003 Already.mp4"
	grows       = "course/01 - Intro/004 Grows.mp4"
	unfinished  = "course/01 - Intro/005 Unfinished.mp4"
	mpeg4       = "course/01 - Intro/006 Mpeg4.mp4"
	h264Codec   = "h264"
	waitStep    = 5 * time.Millisecond
	waitLimit   = 2 * time.Second
)

// fakeVideos stands in for ffprobe, ffmpeg and the library directory, keyed by library path.
type fakeVideos struct {
	mutex    sync.Mutex
	media    map[string]Media
	encodeAs map[string]Media
	blockOn  string
	skips    map[string]Skip
}

func (videos *fakeVideos) Probe(_ context.Context, path string) (Media, error) {
	videos.mutex.Lock()
	defer videos.mutex.Unlock()
	media, ok := videos.media[path]
	if !ok {
		return Media{}, fs.ErrNotExist
	}
	return media, nil
}

func (videos *fakeVideos) Encode(ctx context.Context, source, target string, _ func(time.Duration)) error {
	if source == videos.blockOn {
		<-ctx.Done()
		return ctx.Err()
	}
	videos.mutex.Lock()
	defer videos.mutex.Unlock()
	videos.media[target] = videos.encodeAs[source]
	return nil
}

func (videos *fakeVideos) Rename(oldPath, newPath string) error {
	videos.mutex.Lock()
	defer videos.mutex.Unlock()
	videos.media[newPath] = videos.media[oldPath]
	delete(videos.media, oldPath)
	return nil
}

func (videos *fakeVideos) Remove(path string) error {
	videos.mutex.Lock()
	defer videos.mutex.Unlock()
	if _, ok := videos.media[path]; !ok {
		return fs.ErrNotExist
	}
	delete(videos.media, path)
	return nil
}

func (videos *fakeVideos) Find(_ context.Context, path string) (Skip, bool, error) {
	videos.mutex.Lock()
	defer videos.mutex.Unlock()
	skip, ok := videos.skips[path]
	return skip, ok, nil
}

func (videos *fakeVideos) Save(_ context.Context, skip Skip) error {
	videos.mutex.Lock()
	defer videos.mutex.Unlock()
	videos.skips[skip.Path] = skip
	return nil
}

func (videos *fakeVideos) snapshot() map[string]Media {
	videos.mutex.Lock()
	defer videos.mutex.Unlock()
	return maps.Clone(videos.media)
}

func newFakeVideos() *fakeVideos {
	minute := time.Minute
	return &fakeVideos{
		media: map[string]Media{
			shrinks:     {Codec: h264Codec, Duration: minute, SizeBytes: 1000},
			drifts:      {Codec: h264Codec, Duration: minute, SizeBytes: 1000},
			alreadyH265: {Codec: HEVCCodec, Duration: minute, SizeBytes: 500},
			grows:       {Codec: h264Codec, Duration: minute, SizeBytes: 300},
			mpeg4:       {Codec: "mpeg4", Duration: minute, SizeBytes: 900},
		},
		encodeAs: map[string]Media{
			shrinks: {Codec: HEVCCodec, Duration: minute + 100*time.Millisecond, SizeBytes: 400},
			drifts:  {Codec: HEVCCodec, Duration: minute + 2*time.Second, SizeBytes: 400},
			grows:   {Codec: HEVCCodec, Duration: minute, SizeBytes: 300},
		},
		skips: map[string]Skip{},
	}
}

func startService(t *testing.T, videos *fakeVideos) (*Service, string) {
	t.Helper()
	files := fstest.MapFS{}
	for _, path := range []string{shrinks, drifts, alreadyH265, grows, unfinished, mpeg4} {
		files[path] = &fstest.MapFile{}
	}
	catalog, err := library.NewCatalog(files)
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	service := NewService(Deps{Catalog: catalog, Prober: videos, Encoder: videos, Files: videos, Skips: videos, Logger: slog.New(slog.DiscardHandler)})
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = service.Run(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})
	return service, catalog.Current().Courses[0].ID
}

func waitFor(t *testing.T, service *Service, courseID string, want State) Run {
	t.Helper()
	for deadline := time.Now().Add(waitLimit); time.Now().Before(deadline); time.Sleep(waitStep) {
		if run, _ := service.Status(courseID); run.State == want {
			return run
		}
	}
	run, _ := service.Status(courseID)
	t.Fatalf("run = %+v, want state %s", run, want)
	return Run{}
}

func TestCompressionReplacesOnlyVerifiedSmallerEncodes(t *testing.T) {
	videos := newFakeVideos()
	service, courseID := startService(t, videos)
	if _, err := service.Start(t.Context(), courseID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	run := waitFor(t, service, courseID, StateDone)

	files := videos.snapshot()
	if files[shrinks].Codec != HEVCCodec || files[drifts].Codec != h264Codec || files[grows].Codec != h264Codec || len(files) != 5 {
		t.Errorf("library after compression = %+v; want only the verified smaller encode swapped in and no leftovers", files)
	}
	if run.Total != 3 || run.Finished != 3 || run.BytesBefore != 1000 || run.BytesAfter != 400 || len(run.Skipped) != 2 {
		t.Errorf("run = %+v", run)
	}
	if len(videos.skips) != 2 {
		t.Errorf("remembered as not worth it = %+v, want the drifting and the growing video", videos.skips)
	}
}

func TestPlanLeavesOutVideosItShouldNotEncodeAndRetriesChangedOnes(t *testing.T) {
	videos := newFakeVideos()
	videos.skips[drifts] = Skip{Path: drifts, SizeBytes: 1000}
	service, courseID := startService(t, videos)

	plan, err := service.Plan(t.Context(), courseID)
	if err != nil || len(plan.Candidates) != 2 || plan.NotWorthIt != 1 || plan.AlreadyHEVC != 1 || plan.OtherCodec != 1 || plan.Unreadable != 1 {
		t.Fatalf("plan = %+v, %v; want only the two H.264 videos not yet found not worth it", plan, err)
	}
	videos.media[drifts] = Media{Codec: h264Codec, Duration: time.Minute, SizeBytes: 2000}
	if plan, _ := service.Plan(t.Context(), courseID); len(plan.Candidates) != 3 {
		t.Errorf("plan after the video changed = %+v, want it back", plan)
	}
}

func TestCancellingKeepsFinishedVideosAndDiscardsTheOneInProgress(t *testing.T) {
	videos := newFakeVideos()
	videos.blockOn = drifts
	service, courseID := startService(t, videos)
	if _, err := service.Start(t.Context(), courseID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for deadline := time.Now().Add(waitLimit); time.Now().Before(deadline); time.Sleep(waitStep) {
		if run, _ := service.Status(courseID); run.Current != nil && run.Current.LectureTitle == "Drifts" {
			break
		}
	}
	if err := service.Cancel(courseID); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	run := waitFor(t, service, courseID, StateCancelled)

	files := videos.snapshot()
	if files[shrinks].Codec != HEVCCodec || files[drifts].Codec != h264Codec || len(files) != 5 || run.Finished != 1 {
		t.Errorf("after cancel: library = %+v, run = %+v", files, run)
	}
}
