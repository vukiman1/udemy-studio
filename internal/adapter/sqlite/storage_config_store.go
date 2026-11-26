package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite/gen"
	"github.com/vukiman1/udemy-studio/internal/objectstore"
	"github.com/vukiman1/udemy-studio/internal/platform/sealedbox"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const storageConfigSetting = "storage_config"

type sealedStorageConfig struct {
	URL             string `json:"url"`
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"` //nolint:gosec // reason: the whole row is sealed before it is stored
}

type StorageConfigStore struct {
	queries *gen.Queries
	box     *sealedbox.Box
}

func NewStorageConfigStore(database *sql.DB, box *sealedbox.Box) *StorageConfigStore {
	return &StorageConfigStore{queries: gen.New(database), box: box}
}

func (store *StorageConfigStore) Load(ctx context.Context) (objectstore.Connection, bool, error) {
	sealed, err := store.queries.GetSetting(ctx, storageConfigSetting)
	if errors.Is(err, sql.ErrNoRows) {
		return objectstore.Connection{}, false, nil
	}
	if err != nil {
		return objectstore.Connection{}, false, fmt.Errorf("read setting %s: %w", storageConfigSetting, err)
	}
	plain, err := store.box.Open(sealed, storageConfigSetting)
	if err != nil {
		return objectstore.Connection{}, false, fmt.Errorf("%w: %w", objectstore.ErrUnreadableConfig, err)
	}
	var stored sealedStorageConfig
	if err := json.Unmarshal(plain, &stored); err != nil {
		return objectstore.Connection{}, false, fmt.Errorf("decode setting %s: %w", storageConfigSetting, err)
	}
	location, err := objectstore.ParseURL(stored.URL)
	if err != nil {
		return objectstore.Connection{}, false, fmt.Errorf("%w: %w", objectstore.ErrUnreadableConfig, err)
	}
	credentials := objectstore.Credentials{AccessKeyID: stored.AccessKeyID, SecretAccessKey: secret.New(stored.SecretAccessKey)}
	return objectstore.Connection{Location: location, Credentials: credentials}, true, nil
}

func (store *StorageConfigStore) Save(ctx context.Context, connection objectstore.Connection) error {
	plain, err := json.Marshal(sealedStorageConfig{
		URL:             connection.Location.URL(),
		AccessKeyID:     connection.Credentials.AccessKeyID,
		SecretAccessKey: connection.Credentials.SecretAccessKey.Reveal(),
	})
	if err != nil {
		return fmt.Errorf("encode setting %s: %w", storageConfigSetting, err)
	}
	sealed, err := store.box.Seal(plain, storageConfigSetting)
	if err != nil {
		return err
	}
	if err := store.queries.PutSetting(ctx, gen.PutSettingParams{Key: storageConfigSetting, Value: sealed}); err != nil {
		return fmt.Errorf("save setting %s: %w", storageConfigSetting, err)
	}
	return nil
}

func (store *StorageConfigStore) Delete(ctx context.Context) error {
	if err := store.queries.DeleteSetting(ctx, storageConfigSetting); err != nil {
		return fmt.Errorf("delete setting %s: %w", storageConfigSetting, err)
	}
	return nil
}
