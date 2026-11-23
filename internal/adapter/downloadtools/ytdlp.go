package downloadtools

import (
	"context"
	"io"
	"strconv"
)

// The Python downloader leaves --concurrent-downloads at its default of 1.
const concurrentFragments = 1

type YTDLP struct {
	Binary string
	Output io.Writer
}

// FetchHLS downloads the stream a local playlist describes; yt-dlp keeps .part files so a rerun resumes.
func (tool YTDLP) FetchHLS(ctx context.Context, playlistURL, outputTemplate string) error {
	arguments := []string{
		"--enable-file-urls",
		"--force-generic-extractor",
		"--concurrent-fragments", strconv.Itoa(concurrentFragments),
		"--downloader", "aria2c",
		"--downloader-args", `aria2c:"--disable-ipv6"`,
		"-o", outputTemplate,
		playlistURL,
	}
	return run(ctx, tool.Output, tool.Binary, arguments)
}
