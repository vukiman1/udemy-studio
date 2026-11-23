package udemy

import (
	"context"
	"fmt"
	"io/fs"
	"os"
)

var testdata = os.DirFS("../../testdata")

type fixturePlaylists struct{}

func (fixturePlaylists) FetchPlaylist(_ context.Context, playlistURL string) (string, error) {
	assetID := masterAssetID.FindStringSubmatch(playlistURL)
	if assetID == nil {
		return "", fmt.Errorf("no asset id in %s", playlistURL)
	}
	content, err := fs.ReadFile(testdata, "udemy/hls/index_"+assetID[masterAssetGroup]+".m3u8")
	return string(content), err
}
