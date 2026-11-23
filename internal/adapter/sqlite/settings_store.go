package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite/gen"
)

const selectedChromeProfileKey = "chrome_profile_dir"

type SettingsStore struct {
	queries *gen.Queries
}

func NewSettingsStore(database *sql.DB) *SettingsStore {
	return &SettingsStore{queries: gen.New(database)}
}

func (store *SettingsStore) SelectedProfile(ctx context.Context) (string, bool, error) {
	value, err := store.queries.GetSetting(ctx, selectedChromeProfileKey)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("read setting %s: %w", selectedChromeProfileKey, err)
	}
	return value, true, nil
}

func (store *SettingsStore) SaveSelectedProfile(ctx context.Context, profileDir string) error {
	if err := store.queries.PutSetting(ctx, gen.PutSettingParams{Key: selectedChromeProfileKey, Value: profileDir}); err != nil {
		return fmt.Errorf("save setting %s: %w", selectedChromeProfileKey, err)
	}
	return nil
}
