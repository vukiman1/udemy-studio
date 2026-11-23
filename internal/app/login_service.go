package app

import (
	"database/sql"
	"time"

	"github.com/vukiman1/udemy-studio/internal/adapter/chromecookie"
	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite"
	"github.com/vukiman1/udemy-studio/internal/login"
)

func newLoginService(chromeDataDir string, database *sql.DB) *login.Service {
	return login.NewService(
		chromecookie.NewStore(chromeDataDir, chromecookie.Keychain{}),
		sqlite.NewSettingsStore(database),
		time.Now,
	)
}
