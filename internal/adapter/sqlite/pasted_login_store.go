package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/vukiman1/udemy-studio/internal/adapter/sqlite/gen"
	"github.com/vukiman1/udemy-studio/internal/login"
	"github.com/vukiman1/udemy-studio/internal/platform/sealedbox"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const pastedLoginsSetting = "udemy_pasted_logins"

type sealedPastedLogin struct {
	Host     string    `json:"host"`
	Token    string    `json:"token"` //nolint:gosec // reason: the whole row is sealed before it is stored
	UserName string    `json:"userName"`
	SavedAt  time.Time `json:"savedAt"`
}

// PastedLoginStore keeps the Udemy tokens pasted by hand sealed in one settings row, like the AI keys.
type PastedLoginStore struct {
	queries *gen.Queries
	box     *sealedbox.Box
}

func NewPastedLoginStore(database *sql.DB, box *sealedbox.Box) *PastedLoginStore {
	return &PastedLoginStore{queries: gen.New(database), box: box}
}

func (store *PastedLoginStore) LoadPastedLogins(ctx context.Context) ([]login.PastedLogin, error) {
	sealed, err := store.queries.GetSetting(ctx, pastedLoginsSetting)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read setting %s: %w", pastedLoginsSetting, err)
	}
	plain, err := store.box.Open(sealed, pastedLoginsSetting)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", login.ErrUnreadablePastedLogins, err)
	}
	var stored []sealedPastedLogin
	if err := json.Unmarshal(plain, &stored); err != nil {
		return nil, fmt.Errorf("decode setting %s: %w", pastedLoginsSetting, err)
	}
	logins := make([]login.PastedLogin, len(stored))
	for position, saved := range stored {
		logins[position] = login.PastedLogin{Host: saved.Host, Token: secret.New(saved.Token), UserName: saved.UserName, SavedAt: saved.SavedAt}
	}
	return logins, nil
}

func (store *PastedLoginStore) SavePastedLogins(ctx context.Context, logins []login.PastedLogin) error {
	stored := make([]sealedPastedLogin, len(logins))
	for position, saved := range logins {
		stored[position] = sealedPastedLogin{Host: saved.Host, Token: saved.Token.Reveal(), UserName: saved.UserName, SavedAt: saved.SavedAt}
	}
	plain, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("encode setting %s: %w", pastedLoginsSetting, err)
	}
	sealed, err := store.box.Seal(plain, pastedLoginsSetting)
	if err != nil {
		return err
	}
	if err := store.queries.PutSetting(ctx, gen.PutSettingParams{Key: pastedLoginsSetting, Value: sealed}); err != nil {
		return fmt.Errorf("save setting %s: %w", pastedLoginsSetting, err)
	}
	return nil
}
