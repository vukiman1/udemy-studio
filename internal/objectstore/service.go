package objectstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	probeKey         = ".studio-probe"
	probeContentType = "text/plain"
	checkTimeout     = 20 * time.Second
	keyHintEdge      = 4
	keyHintMask      = "••••••••"
)

var probeBody = []byte("udemy-studio write check")

type View struct {
	IsConfigured    bool
	URL             string
	AccessKeyIDHint string
}

type Draft struct {
	URL             string
	AccessKeyID     string
	SecretAccessKey string
}

type Check struct {
	Access   Access
	HasIndex bool
}

type Service struct {
	store  ConfigStore
	open   BucketOpener
	logger *slog.Logger
}

func NewService(store ConfigStore, open BucketOpener, logger *slog.Logger) *Service {
	return &Service{store: store, open: open, logger: logger}
}

func (service *Service) View(ctx context.Context) (View, error) {
	connection, found, err := service.load(ctx)
	if err != nil || !found {
		return View{}, err
	}
	return View{
		IsConfigured:    true,
		URL:             connection.Location.URL(),
		AccessKeyIDHint: keyHint(connection.Credentials.AccessKeyID),
	}, nil
}

func (service *Service) Save(ctx context.Context, draft Draft) error {
	connection, err := service.connection(ctx, draft)
	if err != nil {
		return err
	}
	if _, err := service.check(ctx, connection); err != nil {
		return err
	}
	return service.store.Save(ctx, connection)
}

func (service *Service) Check(ctx context.Context, draft Draft) (Check, error) {
	connection, err := service.connection(ctx, draft)
	if err != nil {
		return Check{}, err
	}
	return service.check(ctx, connection)
}

func (service *Service) Bucket(ctx context.Context) (Bucket, error) {
	connection, found, err := service.load(ctx)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrNotConfigured
	}
	return service.open(connection), nil
}

func (service *Service) Forget(ctx context.Context) error {
	return service.store.Delete(ctx)
}

func (service *Service) check(ctx context.Context, connection Connection) (Check, error) {
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()
	bucket := service.open(connection)
	hasIndex, err := hasObject(ctx, bucket, IndexKey)
	if err != nil {
		return Check{}, err
	}
	access, err := probeWrite(ctx, bucket)
	if err != nil {
		return Check{}, err
	}
	return Check{Access: access, HasIndex: hasIndex}, nil
}

func hasObject(ctx context.Context, bucket Bucket, key string) (bool, error) {
	object, err := bucket.Open(ctx, key)
	if errors.Is(err, ErrObjectNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := object.Close(); err != nil {
		return false, fmt.Errorf("close %s: %w", key, err)
	}
	return true, nil
}

func probeWrite(ctx context.Context, bucket Bucket) (Access, error) {
	err := bucket.Put(ctx, Upload{Key: probeKey, Body: bytes.NewReader(probeBody), SizeBytes: int64(len(probeBody)), ContentType: probeContentType})
	if errors.Is(err, ErrAccessDenied) {
		return AccessReadOnly, nil
	}
	if err != nil {
		return "", err
	}
	if err := bucket.Delete(ctx, probeKey); err != nil {
		return "", fmt.Errorf("remove the write check %s: %w", probeKey, err)
	}
	return AccessReadWrite, nil
}

func (service *Service) connection(ctx context.Context, draft Draft) (Connection, error) {
	location, err := ParseURL(draft.URL)
	if err != nil {
		return Connection{}, err
	}
	credentials, err := service.credentials(ctx, draft)
	if err != nil {
		return Connection{}, err
	}
	return Connection{Location: location, Credentials: credentials}, nil
}

func (service *Service) credentials(ctx context.Context, draft Draft) (Credentials, error) {
	keyID, secretKey := strings.TrimSpace(draft.AccessKeyID), strings.TrimSpace(draft.SecretAccessKey)
	if keyID != "" && secretKey != "" {
		return Credentials{AccessKeyID: keyID, SecretAccessKey: secret.New(secretKey)}, nil
	}
	if keyID != "" || secretKey != "" {
		return Credentials{}, ErrNoCredentials
	}
	saved, found, err := service.load(ctx)
	if err != nil {
		return Credentials{}, err
	}
	if !found {
		return Credentials{}, ErrNoCredentials
	}
	return saved.Credentials, nil
}

func (service *Service) load(ctx context.Context) (Connection, bool, error) {
	connection, found, err := service.store.Load(ctx)
	if errors.Is(err, ErrUnreadableConfig) {
		service.logger.Warn("starting the storage settings over: the saved ones cannot be read", "err", err)
		return Connection{}, false, nil
	}
	return connection, found, err
}

func keyHint(value string) string {
	if utf8.RuneCountInString(value) <= 2*keyHintEdge {
		return keyHintMask
	}
	return value[:keyHintEdge] + keyHintMask + value[len(value)-keyHintEdge:]
}
