package download

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const (
	playlistFileMode       = 0o600
	videoExtension         = ".mp4"
	ytDLPExtensionTemplate = ".%(ext)s"
	fileURLScheme          = "file"
)

var (
	errNoVideoSource  = errors.New("no downloadable video source")
	errDRMUnsupported = errors.New("DRM-protected lectures are not supported")
	errNoMedia        = errors.New("lecture has no media")
)

func (run *courseDownload) mainFile(ctx context.Context, chapterDir string, lecture udemy.LecturePlan) error {
	if isDownloaded(filepath.Join(chapterDir, lecture.File)) {
		run.summary.AlreadyPresent++
		return nil
	}
	switch lecture.Kind {
	case udemy.KindHTML:
		return run.page(ctx, chapterDir, lecture)
	case udemy.KindVideo:
		return run.video(ctx, chapterDir, lecture)
	case udemy.KindDRM:
		return run.recordFailure(ctx, lecture.Title, errDRMUnsupported)
	case udemy.KindQuiz:
		return run.quiz(ctx, chapterDir, lecture)
	case udemy.KindUnavailable:
		return run.recordFailure(ctx, lecture.Title, errNoMedia)
	}
	return nil
}

// A page with an empty body is left unwritten and uncounted, as in the Python downloader; an article always has its template.
func (run *courseDownload) page(ctx context.Context, chapterDir string, lecture udemy.LecturePlan) error {
	content := lecture.HTMLBody
	if slices.ContainsFunc(lecture.Assets, func(asset udemy.AssetPlan) bool { return asset.Type == udemy.AssetArticle }) {
		content = articlePage(lecture)
	}
	if content == "" {
		return nil
	}
	if err := writeLibraryFile(filepath.Join(chapterDir, lecture.File), []byte(content)); err != nil {
		return run.recordFailure(ctx, lecture.Title, err)
	}
	run.summary.Downloaded++
	return nil
}

func (run *courseDownload) video(ctx context.Context, chapterDir string, lecture udemy.LecturePlan) error {
	if lecture.Video == nil {
		return run.recordFailure(ctx, lecture.Title, errNoVideoSource)
	}
	var err error
	if lecture.Video.Type == udemy.SourceHLS {
		err = run.hls(ctx, chapterDir, lecture)
	} else {
		err = run.tools.Files.FetchFile(ctx, lecture.Video.URL, chapterDir, lecture.File)
	}
	if err != nil {
		return run.recordFailure(ctx, lecture.Title, err)
	}
	run.summary.Downloaded++
	return nil
}

func (run *courseDownload) hls(ctx context.Context, chapterDir string, lecture udemy.LecturePlan) error {
	playlist, err := run.tools.Playlists.FetchPlaylist(ctx, lecture.Video.URL)
	if err != nil {
		return fmt.Errorf("fetch HLS playlist: %w", err)
	}
	playlistPath := filepath.Join(run.options.TempDir, lecture.Video.PlaylistFile)
	if err := os.WriteFile(playlistPath, []byte(playlist), playlistFileMode); err != nil {
		return fmt.Errorf("save HLS playlist: %w", err)
	}
	playlistURL := url.URL{Scheme: fileURLScheme, Path: playlistPath}
	output := filepath.Join(chapterDir, strings.TrimSuffix(lecture.File, videoExtension)+ytDLPExtensionTemplate)
	return run.tools.Streams.FetchHLS(ctx, playlistURL.String(), output)
}
