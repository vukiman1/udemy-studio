package app

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/vukiman1/udemy-studio/internal/adapter/claude"
	"github.com/vukiman1/udemy-studio/internal/adapter/openai"
	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite"
	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/platform/sealedbox"
)

func newAISettings(box *sealedbox.Box, database *sql.DB, logger *slog.Logger) *llm.Service {
	profiles := []llm.Profile{claude.Profile(), openai.Profile()}
	return llm.NewService(sqlite.NewAIConfigStore(database, box), profiles, time.Now, logger)
}
