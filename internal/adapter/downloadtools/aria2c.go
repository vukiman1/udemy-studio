package downloadtools

import (
	"context"
	"io"
)

var aria2cOptions = []string{
	"-j16", "-s20", "-x16", "-c",
	"--auto-file-renaming=false", "--summary-interval=0", "--disable-ipv6", "--follow-torrent=false",
}

type Aria2c struct {
	Binary string
	Output io.Writer
}

// FetchFile downloads url to dir/name, resuming a partial file left by an earlier run.
func (tool Aria2c) FetchFile(ctx context.Context, url, dir, name string) error {
	arguments := append([]string{url, "-o", name, "-d", dir}, aria2cOptions...)
	return run(ctx, tool.Output, tool.Binary, arguments)
}
