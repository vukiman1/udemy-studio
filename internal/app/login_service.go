package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/vukiman1/udemy-studio/internal/adapter/chromecookie"
	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite"
	"github.com/vukiman1/udemy-studio/internal/adapter/udemyapi"
	"github.com/vukiman1/udemy-studio/internal/login"
	"github.com/vukiman1/udemy-studio/internal/platform/sealedbox"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const udemyHostSuffix = ".udemy.com"

func newLoginService(chromeDataDir string, database *sql.DB, box *sealedbox.Box, logger *slog.Logger) *login.Service {
	return login.NewService(login.Deps{
		Cookies:  chromecookie.NewStore(chromeDataDir, chromecookie.Keychain{}),
		Settings: sqlite.NewSettingsStore(database),
		Pasted:   sqlite.NewPastedLoginStore(database, box),
		Checker:  udemyTokenChecker{logger: logger},
		Now:      time.Now,
		Logger:   logger,
	})
}

// udemyTokenChecker asks the portal which account a pasted token logs in as.
type udemyTokenChecker struct {
	logger *slog.Logger
}

func (checker udemyTokenChecker) UserName(ctx context.Context, host string, token secret.Secret) (string, error) {
	client, err := udemyapi.New(strings.TrimSuffix(host, udemyHostSuffix), token, checker.logger)
	if err != nil {
		return "", err
	}
	name, err := client.CurrentUserName(ctx)
	var statusErr *udemyapi.StatusError
	if errors.As(err, &statusErr) && statusErr.IsAuthFailure() {
		return "", fmt.Errorf("%w (HTTP %d)", login.ErrTokenRejected, statusErr.Status)
	}
	return name, err
}
