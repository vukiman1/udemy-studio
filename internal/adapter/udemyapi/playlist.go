package udemyapi

import (
	"context"
	"errors"
	"fmt"

	"github.com/vukiman1/udemy-studio/internal/udemy"
)

func (client *Client) FetchPlaylist(ctx context.Context, playlistURL string) (string, error) {
	playlist, err := client.getText(ctx, playlistURL)
	var statusErr *StatusError
	if errors.As(err, &statusErr) {
		return "", fmt.Errorf("%w: %w", udemy.ErrPlaylistUnavailable, err)
	}
	return playlist, err
}
