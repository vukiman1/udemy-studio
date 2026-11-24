package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite/gen"
	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/platform/sealedbox"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const aiConfigSetting = "ai_config"

type sealedAIConfig struct {
	Active    llm.Provider                    `json:"active"`
	IsEnabled bool                            `json:"enabled"`
	Providers map[llm.Provider]sealedProvider `json:"providers"`
}

type sealedProvider struct {
	APIKey string `json:"apiKey"` //nolint:gosec // reason: the whole row is sealed before it is stored
}

// AIConfigStore keeps the AI settings, keys included, sealed in one settings row; the sealing key lives outside the database.
type AIConfigStore struct {
	queries *gen.Queries
	box     *sealedbox.Box
}

func NewAIConfigStore(database *sql.DB, box *sealedbox.Box) *AIConfigStore {
	return &AIConfigStore{queries: gen.New(database), box: box}
}

func (store *AIConfigStore) Load(ctx context.Context) (llm.Config, bool, error) {
	sealed, err := store.queries.GetSetting(ctx, aiConfigSetting)
	if errors.Is(err, sql.ErrNoRows) {
		return llm.Config{}, false, nil
	}
	if err != nil {
		return llm.Config{}, false, fmt.Errorf("read setting %s: %w", aiConfigSetting, err)
	}
	plain, err := store.box.Open(sealed, aiConfigSetting)
	if err != nil {
		return llm.Config{}, false, fmt.Errorf("%w: %w", llm.ErrUnreadableConfig, err)
	}
	var stored sealedAIConfig
	if err := json.Unmarshal(plain, &stored); err != nil {
		return llm.Config{}, false, fmt.Errorf("decode setting %s: %w", aiConfigSetting, err)
	}
	config := llm.Config{Active: stored.Active, IsEnabled: stored.IsEnabled, Providers: map[llm.Provider]llm.ProviderSettings{}}
	for provider, saved := range stored.Providers {
		config.Providers[provider] = llm.ProviderSettings{APIKey: secret.New(saved.APIKey)}
	}
	return config, true, nil
}

func (store *AIConfigStore) Save(ctx context.Context, config llm.Config) error {
	stored := sealedAIConfig{Active: config.Active, IsEnabled: config.IsEnabled, Providers: map[llm.Provider]sealedProvider{}}
	for provider, settings := range config.Providers {
		stored.Providers[provider] = sealedProvider{APIKey: settings.APIKey.Reveal()}
	}
	plain, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("encode setting %s: %w", aiConfigSetting, err)
	}
	sealed, err := store.box.Seal(plain, aiConfigSetting)
	if err != nil {
		return err
	}
	if err := store.queries.PutSetting(ctx, gen.PutSettingParams{Key: aiConfigSetting, Value: sealed}); err != nil {
		return fmt.Errorf("save setting %s: %w", aiConfigSetting, err)
	}
	return nil
}

func (store *AIConfigStore) Delete(ctx context.Context) error {
	if err := store.queries.DeleteSetting(ctx, aiConfigSetting); err != nil {
		return fmt.Errorf("delete setting %s: %w", aiConfigSetting, err)
	}
	return nil
}
