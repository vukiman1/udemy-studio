package objectstore

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

type Access string

const (
	AccessReadOnly  Access = "read-only"
	AccessReadWrite Access = "read-write"
)

const IndexKey = "studio.json"

var (
	ErrInvalidURL       = errors.New("invalid S3 API URL")
	ErrNoCredentials    = errors.New("enter both the access key ID and the secret access key")
	ErrNotConfigured    = errors.New("no cloud storage is configured")
	ErrAccessDenied     = errors.New("the bucket refused these keys")
	ErrBucketNotFound   = errors.New("the bucket does not exist")
	ErrObjectNotFound   = errors.New("no such object in the bucket")
	ErrUnreachable      = errors.New("the storage could not be reached")
	ErrUnreadableConfig = errors.New("the saved storage settings cannot be read")
)

type Credentials struct {
	AccessKeyID     string
	SecretAccessKey secret.Secret
}

type Connection struct {
	Location    Location
	Credentials Credentials
}

type Object struct {
	Key        string
	SizeBytes  int64
	ModifiedAt time.Time
}

type Upload struct {
	Key         string
	Body        io.ReadSeeker
	SizeBytes   int64
	ContentType string
}

type LinkRequest struct {
	Key                string
	ExpiresIn          time.Duration
	ContentDisposition string
}

type Bucket interface {
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Link(ctx context.Context, request LinkRequest) (string, error)
	Put(ctx context.Context, upload Upload) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]Object, error)
}

type BucketOpener func(connection Connection) Bucket

type ConfigStore interface {
	Load(ctx context.Context) (Connection, bool, error)
	Save(ctx context.Context, connection Connection) error
	Delete(ctx context.Context) error
}
