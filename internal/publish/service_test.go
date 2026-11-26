package publish

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"slices"
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
	introChapter    = "01 - Intro"
	deepChapter     = "02 - Deep"
	firstFileCount  = 6
)

var (
	errUploadBroke = errors.New("upload broke")
	publishedTime  = time.Date(2026, 11, 26, 8, 0, 0, 0, time.UTC)
)

type memoryBucket struct {
	mutex      sync.Mutex
	objects    map[string]objectstore.Object
	bodies     map[string][]byte
	putKeys    []string
	deleteKeys []string
	onPut      func(ctx context.Context, key string) error
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
	bucket.deleteKeys = append(bucket.deleteKeys, key)
	return nil
}

func (*memoryBucket) Link(context.Context, objectstore.LinkRequest) (string, error) {
	return "", nil
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
		firstCourseDir + "/.course.json":                     {Data: []byte(`{"format":1,"title":"First","lectures":[]}`)},
		firstCourseDir + "/cover.jpg":                        {Data: []byte("jpeg")},
		firstCourseDir + "/" + deepChapter + "/003 Deep.mp4": {Data: []byte("video-three")},
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

func allChapters(t *testing.T, service *Service, dir string) []string {
	t.Helper()
	plan, err := service.Plan(t.Context(), courseID(t, service, dir))
	if err != nil {
		t.Fatalf("Plan %s: %v", dir, err)
	}
	var chapters []string
	for _, chapter := range plan.Chapters() {
		chapters = append(chapters, chapter.Dir)
	}
	return chapters
}

func publish(t *testing.T, service *Service, dir string) Run {
	t.Helper()
	return publishChapters(t, service, dir, allChapters(t, service, dir))
}

func publishChapters(t *testing.T, service *Service, dir string, chapters []string) Run {
	t.Helper()
	if _, err := service.Start(t.Context(), courseID(t, service, dir), chapters); err != nil {
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

	if run.State != StateDone || run.SentFiles != firstFileCount {
		t.Fatalf("run = %+v, want done with %d files sent", run, firstFileCount)
	}
	if last := bucket.putKeys[len(bucket.putKeys)-1]; last != objectstore.IndexKey {
		t.Errorf("last write = %q, want the index", last)
	}
	index := bucket.index(t)
	if index.Course.Dir != firstCourseDir || len(index.Files) != firstFileCount || !index.PublishedAt.Equal(publishedTime) {
		t.Errorf("index = %+v, want first-course with %d files", index, firstFileCount)
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
	if run.State != StateDone || run.SentFiles != 0 || run.SkippedFiles != firstFileCount {
		t.Errorf("run = %+v, want nothing sent and %d skipped", run, firstFileCount)
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

	if _, err := service.Start(t.Context(), courseID(t, service, firstCourseDir), allChapters(t, service, firstCourseDir)); !errors.Is(err, ErrTooLarge) {
		t.Errorf("Start error = %v, want ErrTooLarge", err)
	}
	if len(bucket.putKeys) != 0 {
		t.Errorf("bucket was written: %v", bucket.putKeys)
	}
}

func TestPublishingSomeChaptersSendsOnlyThemWithTheCourseFiles(t *testing.T) {
	bucket := newMemoryBucket()
	service := newTestService(t, courseTree(), bucket)
	publish(t, service, firstCourseDir)

	run := publishChapters(t, service, firstCourseDir, []string{deepChapter})
	if run.State != StateDone {
		t.Fatalf("run = %+v, want done", run)
	}
	for _, path := range []string{".course.json", "cover.jpg", deepChapter + "/003 Deep.mp4"} {
		if _, found := bucket.objects[ObjectKey(firstCourseDir, path)]; !found {
			t.Errorf("%s is missing from the bucket", path)
		}
	}
	if _, found := bucket.objects[ObjectKey(firstCourseDir, introChapter+"/001 Welcome.mp4")]; found {
		t.Error("a chapter left out is still in the bucket")
	}
	index := bucket.index(t)
	if len(index.Chapters) != 1 || index.Chapters[0] != deepChapter || index.Course.ChapterCount != 2 {
		t.Errorf("index chapters = %v of %d, want only %q of 2", index.Chapters, index.Course.ChapterCount, deepChapter)
	}
}

func TestAChapterThatFitsCanBePublishedFromACourseOverTheLimit(t *testing.T) {
	bucket := newMemoryBucket()
	tree := courseTree()
	tree[firstCourseDir+"/"+introChapter+"/003 Long.mp4"] = &fstest.MapFile{Data: make([]byte, testLimitBytes)}
	service := newTestService(t, tree, bucket)

	if run := publishChapters(t, service, firstCourseDir, []string{deepChapter}); run.State != StateDone {
		t.Errorf("run = %+v, want done", run)
	}
}

func TestAnUnknownChapterIsRefused(t *testing.T) {
	service := newTestService(t, courseTree(), newMemoryBucket())
	_, err := service.Start(t.Context(), courseID(t, service, firstCourseDir), []string{"99 - Missing"})
	if !errors.Is(err, ErrUnknownChapter) {
		t.Errorf("Start error = %v, want ErrUnknownChapter", err)
	}
}

func TestDroppingAChapterHidesTheIndexBeforeItsFilesGo(t *testing.T) {
	bucket := newMemoryBucket()
	service := newTestService(t, courseTree(), bucket)
	publish(t, service, firstCourseDir)

	publishChapters(t, service, firstCourseDir, []string{deepChapter})
	indexAt := slices.Index(bucket.deleteKeys, objectstore.IndexKey)
	chapterFileAt := slices.Index(bucket.deleteKeys, ObjectKey(firstCourseDir, introChapter+"/001 Welcome.mp4"))
	if indexAt < 0 || chapterFileAt < 0 || indexAt > chapterFileAt {
		t.Errorf("deletes = %v, want the index removed before the chapter's files", bucket.deleteKeys)
	}
}

func TestLeftoversTheIndexDoesNotListKeepItPublished(t *testing.T) {
	bucket := newMemoryBucket()
	service := newTestService(t, courseTree(), bucket)
	publish(t, service, firstCourseDir)
	stray := ObjectKey(firstCourseDir, "01 - Intro/000 Stray.mp4")
	bucket.objects[stray] = objectstore.Object{Key: stray, ModifiedAt: time.Now()}

	publish(t, service, firstCourseDir)
	if slices.Contains(bucket.deleteKeys, objectstore.IndexKey) {
		t.Errorf("deletes = %v, want the index kept while only unlisted leftovers go", bucket.deleteKeys)
	}
	if _, found := bucket.objects[stray]; found {
		t.Error("the leftover is still in the bucket")
	}
}

func changedChapters(t *testing.T, service *Service, dir string) (map[string]bool, bool) {
	t.Helper()
	plan, err := service.Plan(t.Context(), courseID(t, service, dir))
	if err != nil {
		t.Fatalf("Plan %s: %v", dir, err)
	}
	changed := map[string]bool{}
	for _, chapter := range plan.Chapters() {
		changed[chapter.Dir] = chapter.HasChanges
	}
	return changed, plan.SharedChanged()
}

func TestThePlanSeesNothingToSendRightAfterPublishing(t *testing.T) {
	service := newTestService(t, courseTree(), newMemoryBucket())
	publish(t, service, firstCourseDir)

	changed, shared := changedChapters(t, service, firstCourseDir)
	if changed[introChapter] || changed[deepChapter] || shared {
		t.Errorf("changes = %v, shared = %v; want none", changed, shared)
	}
}

func TestThePlanMarksTheChapterWhoseFilesChangedOrVanished(t *testing.T) {
	tree := courseTree()
	service := newTestService(t, tree, newMemoryBucket())
	publish(t, service, firstCourseDir)
	tree[firstCourseDir+"/"+deepChapter+"/003 Deep.mp4"] = &fstest.MapFile{Data: []byte("compressed"), ModTime: time.Now().Add(time.Hour)}
	delete(tree, firstCourseDir+"/"+introChapter+"/001 Welcome_en.vtt")

	changed, shared := changedChapters(t, service, firstCourseDir)
	if !changed[deepChapter] || !changed[introChapter] || shared {
		t.Errorf("changes = %v, shared = %v; want both chapters and nothing shared", changed, shared)
	}
}

func TestThePlanMarksSharedChangesForACourseFile(t *testing.T) {
	tree := courseTree()
	service := newTestService(t, tree, newMemoryBucket())
	publish(t, service, firstCourseDir)
	delete(tree, firstCourseDir+"/cover.jpg")

	if _, shared := changedChapters(t, service, firstCourseDir); !shared {
		t.Error("a removed cover is not marked as a shared change")
	}
}
