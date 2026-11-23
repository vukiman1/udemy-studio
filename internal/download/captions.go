package download

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vukiman1/udemy-studio/internal/subtitle"
	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const (
	captionAttempts = 4
	vttExtension    = ".vtt"
)

func (run *courseDownload) captions(ctx context.Context, chapterDir string, lecture udemy.LecturePlan) error {
	for _, caption := range lecture.Captions {
		if err := run.caption(ctx, chapterDir, caption); err != nil {
			return run.recordFailure(ctx, fmt.Sprintf("%s (caption: %s)", lecture.Title, caption.Language), err)
		}
	}
	return nil
}

// A VTT left without its SRT, by a failed write or an interrupt, is converted without downloading it again.
func (run *courseDownload) caption(ctx context.Context, chapterDir string, caption udemy.CaptionPlan) error {
	captionPath, srtPath := filepath.Join(chapterDir, caption.File), filepath.Join(chapterDir, caption.SRTFile)
	isVTT := strings.HasSuffix(caption.File, vttExtension)
	if isVTT && isDownloaded(srtPath) {
		return nil
	}
	if !isDownloaded(captionPath) {
		if err := run.fetchCaption(ctx, chapterDir, caption); err != nil {
			return err
		}
	}
	if !isVTT {
		return nil
	}
	return convertCaption(captionPath, srtPath)
}

// fetchCaption retries a failed download, rechecking the file first as the Python downloader's recursion does.
func (run *courseDownload) fetchCaption(ctx context.Context, chapterDir string, caption udemy.CaptionPlan) error {
	for attempt := 1; ; attempt++ {
		err := run.tools.Files.FetchFile(ctx, caption.URL, chapterDir, caption.File)
		if err == nil || ctx.Err() != nil || attempt == captionAttempts {
			return err
		}
		run.logger.Warn("caption download failed, retrying", "caption", caption.File, "attempt", attempt, "err", err)
		if isDownloaded(filepath.Join(chapterDir, caption.File)) {
			return nil
		}
	}
}

// A VTT that does not convert is removed, so the next run downloads it again instead of finding it present.
func convertCaption(vttPath, srtPath string) error {
	vtt, err := os.ReadFile(vttPath) //nolint:gosec // reason: the path is the caption this download just wrote
	if err != nil {
		return fmt.Errorf("read %s: %w", filepath.Base(vttPath), err)
	}
	srt, err := subtitle.ConvertVTTToSRT(vtt)
	if err != nil {
		return errors.Join(fmt.Errorf("convert %s: %w", filepath.Base(vttPath), err), os.Remove(vttPath))
	}
	if err := writeLibraryFile(srtPath, srt); err != nil {
		return fmt.Errorf("write %s: %w", filepath.Base(srtPath), err)
	}
	return os.Remove(vttPath)
}
