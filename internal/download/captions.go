package download

import (
	"context"
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

// Each retry rechecks the files first, as Python's recursion does, so a partial file from a failed attempt counts as done.
func (run *courseDownload) caption(ctx context.Context, chapterDir string, caption udemy.CaptionPlan) error {
	isVTT := strings.HasSuffix(caption.File, vttExtension)
	for attempt := 1; ; attempt++ {
		if isFile(filepath.Join(chapterDir, caption.File)) || (isVTT && isFile(filepath.Join(chapterDir, caption.SRTFile))) {
			return nil
		}
		err := run.tools.Files.FetchFile(ctx, caption.URL, chapterDir, caption.File)
		if err == nil {
			break
		}
		if ctx.Err() != nil || attempt == captionAttempts {
			return err
		}
		run.logger.Warn("caption download failed, retrying", "caption", caption.File, "attempt", attempt, "err", err)
	}
	if !isVTT {
		return nil
	}
	return convertCaption(filepath.Join(chapterDir, caption.File), filepath.Join(chapterDir, caption.SRTFile))
}

// The Python downloader creates the SRT before converting, so a failed conversion leaves it empty and the VTT in place.
func convertCaption(vttPath, srtPath string) error {
	if err := os.WriteFile(srtPath, nil, fileMode); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Base(srtPath), err)
	}
	vtt, err := os.ReadFile(vttPath) //nolint:gosec // reason: the path is the caption this download just wrote
	if err != nil {
		return fmt.Errorf("read %s: %w", filepath.Base(vttPath), err)
	}
	srt, err := subtitle.ConvertVTTToSRT(vtt)
	if err != nil {
		return fmt.Errorf("convert %s: %w", filepath.Base(vttPath), err)
	}
	if err := os.WriteFile(srtPath, srt, fileMode); err != nil { //nolint:gosec // reason: the path is built from the plan's single-element names under the chapter directory
		return fmt.Errorf("write %s: %w", filepath.Base(srtPath), err)
	}
	return os.Remove(vttPath)
}
