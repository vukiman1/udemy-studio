package publish

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/objectstore"
)

const (
	firstCourseDir  = "first-course"
	secondCourseDir = "second-course"
	waitLimit       = 5 * time.Second
	waitStep        = 10 * time.Millisecond
	testLimitBytes  = 1_000
)

var (
	errUploadBroke = errors.New("upload broke")
	publishedTime  = time.Date(2026, 11, 26, 8, 0, 0, 0, time.UTC)
)

type memoryBucket struct {
	mutex   sync.Mutex
	objects map[string]objectstore.Object
	bodies  map[string][]byte
	putKeys []string
	onPut   func(ctx context.Context, key string) error
}

func newMemoryBucket() *memoryBucket {
	return &memoryBucket{objects: map[string]objectstore.Object{}, bodies: map[string][]byte{}}
}

func (bucket *memoryBucket) Bucket(context.Context) (objectstore.Bucket, error) {
	return bucket, nil
}

func (bucket *memoryBucket) Open(_ context.Context, key string) (io.ReadCloser, error) {
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()
	body, found := bucket.bodies[key]
	if !found {
		return nil, objectstore.ErrObjectNotFound
	}
	return io.NopCloser(strings.NewReader(string(body))), nil
}

func (bucket *memoryBucket) Put(ctx context.Context, upload objectstore.Upload) error {
	if bucket.onPut != nil {
		if err := bucket.onPut(ctx, upload.Key); err != nil {
			return err
		}
	}
	body, err := io.ReadAll(upload.Body)
	if err != nil {
		return err
	}
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()
	bucket.objects[upload.Key] = objectstore.Object{Key: upload.Key, SizeBytes: int64(len(body)), ModifiedAt: time.Now()}
	bucket.bodies[upload.Key] = body
	bucket.putKeys = append(bucket.putKeys, upload.Key)
	return nil
}

func (bucket *memoryBucket) Delete(_ context.Context, key string) error {
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()
	delete(bucket.objects, key)
	delete(bucket.bodies, key)
	return nil
}

func (bucket *memoryBucket) List(_ context.Context, prefix string) ([]objectstore.Object, error) {
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()
	var listed []objectstore.Object
	for key, object := range bucket.objects {
		if strings.HasPrefix(key, prefix) {
			listed = append(listed, object)
		}
	}
	return listed, nil
}

func (bucket *memoryBucket) index(t *testing.T) Index {
	t.Helper()
	index, found, err := ReadIndex(t.Context(), bucket)
	if err != nil || !found {
		t.Fatalf("ReadIndex = %v, %v", found, err)
	}
	return index
}

func courseTree() fstest.MapFS {
	return fstest.MapFS{
		firstCourseDir + "/01 - Intro/001 Welcome.mp4":       {Data: []byte("video-one")},
		firstCourseDir + "/01 - Intro/001 Welcome_en.vtt":    {Data: []byte("WEBVTT")},
		firstCourseDir + "/01 - Intro/002 Setup.mp4":         {Data: []byte("video-two")},
		firstCourseDir + "/01 - Intro/002 Setup.mp4.aria2":   {Data: []byte("partial")},
		firstCourseDir + "/.DS_Store":                        {Data: []byte("finder")},
		firstCourseDir + "/.cache/thumb.jpg":                 {Data: []byte("cache")},
		secondCourseDir + "/01 - Basics/001 Hello.mp4":       {Data: []byte("hello")},
		secondCourseDir + "/01 - Basics/001 Hello_en.vtt":    {Data: []byte("WEBVTT")},
		secondCourseDir + "/01 - Basics/001 Hello.mp4.aria2": {Data: []byte("partial")},
	}
}

func newTestService(t *testing.T, tree fstest.MapFS, bucket *memoryBucket) *Service {
	t.Helper()
	catalog, err := library.NewCatalog(tree)
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	service := NewService(Deps{
		Catalog:    catalog,
		Files:      tree,
		Buckets:    bucket,
		LimitBytes: testLimitBytes,
		Now:        func() time.Time { return publishedTime },
		Logger:     slog.New(slog.DiscardHandler),
	})
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	go func() { _ = service.Run(ctx) }()
	return service
}

func courseID(t *testing.T, service *Service, dir string) string {
	t.Helper()
	for _, course := range service.deps.Catalog.Current().Courses {
		if course.Dir == dir {
			return course.ID
		}
	}
	t.Fatalf("no course in %s", dir)
	return ""
}

func publish(t *testing.T, service *Service, dir string) Run {
	t.Helper()
	if _, err := service.Start(t.Context(), courseID(t, service, dir)); err != nil {
		t.Fatalf("Start %s: %v", dir, err)
	}
	deadline := time.Now().Add(waitLimit)
	for time.Now().Before(deadline) {
		if run, _ := service.Status(); run.State != StateRunning {
			return run
		}
		time.Sleep(waitStep)
	}
	t.Fatalf("publishing %s did not finish", dir)
	return Run{}
}

func TestPublishSendsTheFinishedFilesAndWritesTheIndexLast(t *testing.T) {
	bucket := newMemoryBucket()
	run := publish(t, newTestService(t, courseTree(), bucket), firstCourseDir)

	if run.State != StateDone || run.SentFiles != 3 {
		t.Fatalf("run = %+v, want done with 3 files sent", run)
	}
	if last := bucket.putKeys[len(bucket.putKeys)-1]; last != objectstore.IndexKey {
		t.Errorf("last write = %q, want the index", last)
	}
	index := bucket.index(t)
	if index.Course.Dir != firstCourseDir || len(index.Files) != 3 || !index.PublishedAt.Equal(publishedTime) {
		t.Errorf("index = %+v, want first-course with 3 files", index)
	}
	if _, found := bucket.objects[ObjectKey(firstCourseDir, "01 - Intro/002 Setup.mp4.aria2")]; found {
		t.Error("an unfinished download was published")
	}
}

func TestPublishingAgainSendsNothingThatIsAlreadyThere(t *testing.T) {
	bucket := newMemoryBucket()
	service := newTestService(t, courseTree(), bucket)
	publish(t, service, firstCourseDir)

	run := publish(t, service, firstCourseDir)
	if run.State != StateDone || run.SentFiles != 0 || run.SkippedFiles != 3 {
		t.Errorf("run = %+v, want nothing sent and 3 skipped", run)
	}
}

func TestAFailedRepublishKeepsThePublishedIndex(t *testing.T) {
	bucket := newMemoryBucket()
	tree := courseTree()
	service := newTestService(t, tree, bucket)
	publish(t, service, firstCourseDir)
	before, err := json.Marshal(bucket.index(t))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	tree[firstCourseDir+"/01 - Intro/002 Setup.mp4"] = &fstest.MapFile{Data: []byte("video-two-compressed"), ModTime: time.Now().Add(time.Hour)}
	bucket.onPut = func(context.Context, string) error { return errUploadBroke }

	run := publish(t, service, firstCourseDir)
	after, err := json.Marshal(bucket.index(t))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if run.State != StateFailed || !strings.Contains(run.Failure, errUploadBroke.Error()) {
		t.Errorf("run = %+v, want failed with the upload error", run)
	}
	if string(after) != string(before) {
		t.Errorf("index changed after a failed publish:\nbefore %s\nafter  %s", before, after)
	}
}

func TestCancellingStopsBeforeTheIndexIsWritten(t *testing.T) {
	bucket := newMemoryBucket()
	service := newTestService(t, courseTree(), bucket)
	bucket.onPut = func(ctx context.Context, _ string) error {
		if err := service.Cancel(); err != nil {
			return err
		}
		<-ctx.Done()
		return ctx.Err()
	}

	run := publish(t, service, firstCourseDir)
	if run.State != StateCancelled {
		t.Errorf("run = %+v, want cancelled", run)
	}
	if _, found := bucket.objects[objectstore.IndexKey]; found {
		t.Error("an index was written for a cancelled publish")
	}
}

func TestPublishingAnotherCourseRemovesTheOldOne(t *testing.T) {
	bucket := newMemoryBucket()
	service := newTestService(t, courseTree(), bucket)
	publish(t, service, firstCourseDir)

	publish(t, service, secondCourseDir)
	for key := range bucket.objects {
		if strings.HasPrefix(key, CoursesPrefix+firstCourseDir+"/") {
			t.Errorf("%s is still in the bucket", key)
		}
	}
	if index := bucket.index(t); index.Course.Dir != secondCourseDir {
		t.Errorf("index course = %q, want %q", index.Course.Dir, secondCourseDir)
	}
}

func TestACourseOverTheLimitIsRefused(t *testing.T) {
	bucket := newMemoryBucket()
	tree := courseTree()
	tree[firstCourseDir+"/01 - Intro/003 Long.mp4"] = &fstest.MapFile{Data: make([]byte, testLimitBytes)}
	service := newTestService(t, tree, bucket)

	if _, err := service.Start(t.Context(), courseID(t, service, firstCourseDir)); !errors.Is(err, ErrTooLarge) {
		t.Errorf("Start error = %v, want ErrTooLarge", err)
	}
	if len(bucket.putKeys) != 0 {
		t.Errorf("bucket was written: %v", bucket.putKeys)
	}
}
