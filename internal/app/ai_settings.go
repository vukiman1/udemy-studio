package app

import (
	"database/sql"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/vukiman1/udemy-studio/internal/adapter/claude"
	"github.com/vukiman1/udemy-studio/internal/adapter/openai"
	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite"
	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/platform/sealedbox"
)

// The key that seals settings such as the API key lives beside the database, never inside it.
const settingsKeyFileName = "settings-encryption.key"

func newAISettings(dataDir string, database *sql.DB, logger *slog.Logger) (*llm.Service, error) {
	box, err := sealedbox.LoadOrCreate(filepath.Join(dataDir, settingsKeyFileName))
	if err != nil {
		return nil, err
	}
	profiles := []llm.Profile{claude.Profile(), openai.Profile()}
	return llm.NewService(sqlite.NewAIConfigStore(database, box), profiles, time.Now, logger), nil
}
