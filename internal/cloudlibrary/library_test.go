package cloudlibrary

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"testing"
	"testing/fstest"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/objectstore"
	"github.com/vukiman1/udemy-studio/internal/publish"
)

const (
	courseDir    = "demo-course"
	videoPath    = "01 - Intro/001 Welcome.mp4"
	subtitlePath = "01 - Intro/001 Welcome_en.vtt"
	subtitleBody = "WEBVTT\n\n00:00.000 --> 00:01.000\nHello\n"
)

var firstPublish = time.Date(2026, 11, 26, 8, 0, 0, 0, time.UTC)

type bucketSource struct {
	bucket objectstore.Bucket
	err    error
}

func (source bucketSource) Bucket(context.Context) (objectstore.Bucket, error) {
	return source.bucket, source.err
}

type objectsBucket map[string][]byte

func (bucket objectsBucket) Open(_ context.Context, key string) (io.ReadCloser, error) {
	body, found := bucket[key]
	if !found {
		return nil, objectstore.ErrObjectNotFound
	}
	return io.NopCloser(bytes.NewReader(body)), nil
}

func (objectsBucket) Link(_ context.Context, request objectstore.LinkRequest) (string, error) {
	return "https://bucket.example/" + request.Key, nil
}

func (objectsBucket) Put(context.Context, objectstore.Upload) error { return nil }
func (objectsBucket) Delete(context.Context, string) error          { return nil }

func (objectsBucket) List(context.Context, string) ([]objectstore.Object, error) {
	return nil, nil
}

func publishedBucket(t *testing.T, publishedAt time.Time, videoSize int64) objectsBucket {
	t.Helper()
	index := publish.Index{
		Format:      publish.IndexFormat,
		PublishedAt: publishedAt,
		Course:      publish.IndexCourse{Dir: courseDir, Title: "Demo"},
		Files: []publish.IndexFile{
			{Path: videoPath, SizeBytes: videoSize},
			{Path: subtitlePath, SizeBytes: int64(len(subtitleBody))},
		},
	}
	raw, err := json.Marshal(index)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	return objectsBucket{
		objectstore.IndexKey:                       raw,
		publish.ObjectKey(courseDir, videoPath):    make([]byte, videoSize),
		publish.ObjectKey(courseDir, subtitlePath): []byte(subtitleBody),
	}
}

func refreshed(t *testing.T, source BucketSource) *Library {
	t.Helper()
	cloud := New(source, slog.New(slog.DiscardHandler))
	if _, err := cloud.Refresh(t.Context()); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	return cloud
}

func TestThePublishedIndexReadsAsALibrary(t *testing.T) {
	cloud := refreshed(t, bucketSource{bucket: publishedBucket(t, firstPublish, 4)})

	if err := fstest.TestFS(cloud.FS(), courseDir+"/"+videoPath, courseDir+"/"+subtitlePath); err != nil {
		t.Fatalf("TestFS: %v", err)
	}
	catalog, err := library.NewCatalog(cloud.FS())
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	courses := catalog.Current().Courses
	if len(courses) != 1 || len(courses[0].Chapters) != 1 || courses[0].Chapters[0].Lectures[0].VideoPath != courseDir+"/"+videoPath {
		t.Errorf("courses = %+v, want the demo course with its video", courses)
	}
}

func TestVideosAreLeftToSignedLinks(t *testing.T) {
	cloud := refreshed(t, bucketSource{bucket: publishedBucket(t, firstPublish, maxProxyBytes+1)})

	if _, err := cloud.FS().Open(courseDir + "/" + videoPath); !errors.Is(err, ErrTooLargeToProxy) {
		t.Errorf("Open video error = %v, want ErrTooLargeToProxy", err)
	}
	link, err := cloud.Link(t.Context(), courseDir+"/"+videoPath, "")
	if err != nil || link != "https://bucket.example/"+publish.ObjectKey(courseDir, videoPath) {
		t.Errorf("Link = %q, %v; want a link to the video's key", link, err)
	}
}

func TestRefreshReportsAChangeOnlyWhenTheIndexChanges(t *testing.T) {
	source := &bucketSource{bucket: publishedBucket(t, firstPublish, 4)}
	cloud := New(source, slog.New(slog.DiscardHandler))
	changes := []bool{}
	for _, publishedAt := range []time.Time{firstPublish, firstPublish, firstPublish.Add(time.Hour)} {
		source.bucket = publishedBucket(t, publishedAt, 4)
		changed, err := cloud.Refresh(t.Context())
		if err != nil {
			t.Fatalf("Refresh: %v", err)
		}
		changes = append(changes, changed)
	}
	if changes[0] != true || changes[1] != false || changes[2] != true {
		t.Errorf("changes = %v, want [true false true]", changes)
	}
}

func TestNoStorageYetMeansAnEmptyLibrary(t *testing.T) {
	cloud := refreshed(t, bucketSource{err: objectstore.ErrNotConfigured})

	entries, err := fs.ReadDir(cloud.FS(), ".")
	if err != nil || len(entries) != 0 {
		t.Errorf("ReadDir = %v, %v; want an empty library", entries, err)
	}
	if _, err := cloud.Link(t.Context(), courseDir+"/"+videoPath, ""); !errors.Is(err, ErrNoBucket) {
		t.Errorf("Link error = %v, want ErrNoBucket", err)
	}
}
