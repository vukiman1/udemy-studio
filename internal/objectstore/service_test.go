package objectstore_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite"
	"github.com/vukiman1/udemy-studio/internal/objectstore"
	"github.com/vukiman1/udemy-studio/internal/platform/sealedbox"
)

const (
	bucketURL   = "https://abc123.r2.cloudflarestorage.com/studio-dev"
	savedKeyID  = "0123456789abcdef0123456789abcdef"
	savedSecret = "saved-secret-not-real" //nolint:gosec // reason: fake key for tests
)

type fakeBucket struct {
	objects      map[string][]byte
	isReadOnly   bool
	isReadDenied bool
	seenKeyIDs   []string
}

func (bucket *fakeBucket) opener(connection objectstore.Connection) objectstore.Bucket {
	bucket.seenKeyIDs = append(bucket.seenKeyIDs, connection.Credentials.AccessKeyID)
	return bucket
}

func (bucket *fakeBucket) Open(_ context.Context, key string) (io.ReadCloser, error) {
	if bucket.isReadDenied {
		return nil, objectstore.ErrAccessDenied
	}
	body, found := bucket.objects[key]
	if !found {
		return nil, objectstore.ErrObjectNotFound
	}
	return io.NopCloser(bytes.NewReader(body)), nil
}

func (bucket *fakeBucket) Put(_ context.Context, key string, body []byte) error {
	if bucket.isReadOnly {
		return objectstore.ErrAccessDenied
	}
	bucket.objects[key] = body
	return nil
}

func (bucket *fakeBucket) Delete(_ context.Context, key string) error {
	delete(bucket.objects, key)
	return nil
}

func newService(t *testing.T, bucket *fakeBucket) *objectstore.Service {
	t.Helper()
	database, err := sqlite.Open(t.Context(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	box, err := sealedbox.LoadOrCreate(filepath.Join(t.TempDir(), "settings.key"))
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	return objectstore.NewService(sqlite.NewStorageConfigStore(database, box), bucket.opener, slog.New(slog.DiscardHandler))
}

func typedDraft() objectstore.Draft {
	return objectstore.Draft{URL: bucketURL, AccessKeyID: savedKeyID, SecretAccessKey: savedSecret}
}

func TestCheckReportsReadWriteAndLeavesNoProbeBehind(t *testing.T) {
	bucket := &fakeBucket{objects: map[string][]byte{objectstore.IndexKey: []byte("{}")}}
	check, err := newService(t, bucket).Check(t.Context(), typedDraft())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if check != (objectstore.Check{Access: objectstore.AccessReadWrite, HasIndex: true}) {
		t.Errorf("Check = %+v, want read-write with an index", check)
	}
	if len(bucket.objects) != 1 {
		t.Errorf("bucket holds %d objects after the check, want only the index", len(bucket.objects))
	}
}

func TestCheckReportsReadOnlyWhenWritesAreRefused(t *testing.T) {
	bucket := &fakeBucket{objects: map[string][]byte{}, isReadOnly: true}
	check, err := newService(t, bucket).Check(t.Context(), typedDraft())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if check != (objectstore.Check{Access: objectstore.AccessReadOnly}) {
		t.Errorf("Check = %+v, want read-only without an index", check)
	}
}

func TestSaveKeepsNothingWhenTheBucketRefusesToBeRead(t *testing.T) {
	service := newService(t, &fakeBucket{objects: map[string][]byte{}, isReadDenied: true})
	if err := service.Save(t.Context(), typedDraft()); !errors.Is(err, objectstore.ErrAccessDenied) {
		t.Fatalf("Save error = %v, want ErrAccessDenied", err)
	}
	view, err := service.View(t.Context())
	if err != nil {
		t.Fatalf("View: %v", err)
	}
	if view.IsConfigured {
		t.Errorf("View = %+v, want nothing saved", view)
	}
}

func TestEmptyKeysReuseTheSavedOnes(t *testing.T) {
	bucket := &fakeBucket{objects: map[string][]byte{}}
	service := newService(t, bucket)
	if err := service.Save(t.Context(), typedDraft()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := service.Check(t.Context(), objectstore.Draft{URL: bucketURL}); err != nil {
		t.Fatalf("Check with saved keys: %v", err)
	}
	if got := bucket.seenKeyIDs[len(bucket.seenKeyIDs)-1]; got != savedKeyID {
		t.Errorf("check used key %q, want the saved %q", got, savedKeyID)
	}
	view, err := service.View(t.Context())
	if err != nil {
		t.Fatalf("View: %v", err)
	}
	want := objectstore.View{IsConfigured: true, URL: bucketURL, AccessKeyIDHint: "0123••••••••cdef"}
	if view != want {
		t.Errorf("View = %+v, want %+v", view, want)
	}
}

func TestOneKeyWithoutTheOtherIsRefused(t *testing.T) {
	service := newService(t, &fakeBucket{objects: map[string][]byte{}})
	_, err := service.Check(t.Context(), objectstore.Draft{URL: bucketURL, AccessKeyID: savedKeyID})
	if !errors.Is(err, objectstore.ErrNoCredentials) {
		t.Errorf("Check error = %v, want ErrNoCredentials", err)
	}
}
