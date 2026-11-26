package s3bucket

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/objectstore"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	testURLVariable       = "UDEMY_STUDIO_TEST_S3_URL"
	testKeyIDVariable     = "UDEMY_STUDIO_TEST_S3_KEY_ID"
	testSecretVariable    = "UDEMY_STUDIO_TEST_S3_SECRET" //nolint:gosec // reason: the name of an environment variable, not a credential
	readerKeyIDVariable   = "UDEMY_STUDIO_TEST_S3_READER_KEY_ID"
	readerSecretVariable  = "UDEMY_STUDIO_TEST_S3_READER_SECRET" //nolint:gosec // reason: the name of an environment variable, not a credential
	testObjectPrefix      = ".studio-test/"
	testObjectKey         = testObjectPrefix + "demo/001 Welcome.vtt"
	testContentType       = "text/vtt"
	wrongSecretAccessKey  = "not-the-secret" //nolint:gosec // reason: deliberately wrong key
	testObjectBodyContent = "WEBVTT"
	testLinkLifetime      = time.Minute
)

func testConnection(t *testing.T) objectstore.Connection {
	t.Helper()
	rawURL, keyID, secretKey := os.Getenv(testURLVariable), os.Getenv(testKeyIDVariable), os.Getenv(testSecretVariable)
	if rawURL == "" {
		t.Skipf("set %s, %s and %s to an existing bucket and read-write keys", testURLVariable, testKeyIDVariable, testSecretVariable)
	}
	location, err := objectstore.ParseURL(rawURL)
	if err != nil {
		t.Fatalf("ParseURL: %v", err)
	}
	connection := objectstore.Connection{
		Location:    location,
		Credentials: objectstore.Credentials{AccessKeyID: keyID, SecretAccessKey: secret.New(secretKey)},
	}
	return connection
}

func testUpload() objectstore.Upload {
	return objectstore.Upload{
		Key:         testObjectKey,
		Body:        strings.NewReader(testObjectBodyContent),
		SizeBytes:   int64(len(testObjectBodyContent)),
		ContentType: testContentType,
	}
}

func TestObjectsRoundTripThroughTheBucket(t *testing.T) {
	bucket := Open(testConnection(t))
	if _, err := bucket.Open(t.Context(), testObjectKey); !errors.Is(err, objectstore.ErrObjectNotFound) {
		t.Fatalf("Open before Put error = %v, want ErrObjectNotFound", err)
	}
	if err := bucket.Put(t.Context(), testUpload()); err != nil {
		t.Fatalf("Put: %v", err)
	}
	listed, err := bucket.List(t.Context(), testObjectPrefix)
	if err != nil || len(listed) != 1 || listed[0].Key != testObjectKey || listed[0].SizeBytes != int64(len(testObjectBodyContent)) {
		t.Errorf("List = %+v, %v; want only %q", listed, err, testObjectKey)
	}
	object, err := bucket.Open(t.Context(), testObjectKey)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	body, err := io.ReadAll(object)
	_ = object.Close()
	if err != nil || string(body) != testObjectBodyContent {
		t.Errorf("read back %q, %v; want %q", body, err, testObjectBodyContent)
	}
	if err := bucket.Delete(t.Context(), testObjectKey); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestAWrongSecretIsReportedAsAccessDenied(t *testing.T) {
	connection := testConnection(t)
	connection.Credentials.SecretAccessKey = secret.New(wrongSecretAccessKey)
	if _, err := Open(connection).Open(t.Context(), objectstore.IndexKey); !errors.Is(err, objectstore.ErrAccessDenied) {
		t.Errorf("Open error = %v, want ErrAccessDenied", err)
	}
}

func TestAReadOnlyKeyIsDeniedWrites(t *testing.T) {
	connection := testConnection(t)
	keyID, secretKey := os.Getenv(readerKeyIDVariable), os.Getenv(readerSecretVariable)
	if keyID == "" {
		t.Skipf("set %s and %s to read-only keys for the same bucket", readerKeyIDVariable, readerSecretVariable)
	}
	connection.Credentials = objectstore.Credentials{AccessKeyID: keyID, SecretAccessKey: secret.New(secretKey)}
	if err := Open(connection).Put(t.Context(), testUpload()); !errors.Is(err, objectstore.ErrAccessDenied) {
		t.Errorf("Put error = %v, want ErrAccessDenied", err)
	}
}

func TestASignedLinkServesTheObjectWithoutKeys(t *testing.T) {
	bucket := Open(testConnection(t))
	if err := bucket.Put(t.Context(), testUpload()); err != nil {
		t.Fatalf("Put: %v", err)
	}
	t.Cleanup(func() { _ = bucket.Delete(context.Background(), testObjectKey) })
	link, err := bucket.Link(t.Context(), objectstore.LinkRequest{Key: testObjectKey, ExpiresIn: testLinkLifetime})
	if err != nil {
		t.Fatalf("Link: %v", err)
	}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, link, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	request.Header.Set("Range", "bytes=0-1")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("GET link: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil || response.StatusCode != http.StatusPartialContent || string(body) != testObjectBodyContent[:2] {
		t.Errorf("GET link = %d %q, %v; want 206 %q", response.StatusCode, body, err, testObjectBodyContent[:2])
	}
}
