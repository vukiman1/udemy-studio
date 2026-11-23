package udemyapi

import "context"

func (client *Client) FetchPlaylist(ctx context.Context, playlistURL string) (string, error) {
	return client.getText(ctx, playlistURL)
}
