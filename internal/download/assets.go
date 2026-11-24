package download

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const (
	titlePlaceholder         = "__title_placeholder__"
	dataPlaceholder          = "__data_placeholder__"
	lectureNumberPrefixRunes = 4
	internetShortcutFormat   = "[InternetShortcut]\nURL=%s"
	externalLinksFile        = "external-links.txt"
	externalLinkEntryFormat  = "\n%s\n%s\n"
)

//go:embed templates/article_template.html
var articleTemplate string

var lineTerminator = regexp.MustCompile(`\r\n|\r|\n`)

func (run *courseDownload) assets(ctx context.Context, chapterDir string, lecture udemy.LecturePlan) error {
	for _, asset := range lecture.Assets {
		switch asset.Type {
		case udemy.AssetArticle:
			// Written as the lecture's page, once, instead of over it on every run like the Python downloader.
		case udemy.AssetExternalLink:
			if err := writeExternalLink(chapterDir, asset); err != nil {
				return err
			}
		case udemy.AssetEbook, udemy.AssetFile, udemy.AssetPresentation, udemy.AssetAudio, udemy.AssetSourceCode:
			if err := run.attachment(ctx, chapterDir, lecture, asset); err != nil {
				return err
			}
		}
	}
	return nil
}

// Attachments are fetched on every run without checking for the file; aria2c's -c finishes or keeps what is there.
func (run *courseDownload) attachment(ctx context.Context, chapterDir string, lecture udemy.LecturePlan, asset udemy.AssetPlan) error {
	if err := run.tools.Files.FetchFile(ctx, asset.URL, chapterDir, asset.File); err != nil {
		return run.recordFailure(ctx, fmt.Sprintf("%s (asset: %s)", lecture.Title, asset.File), err)
	}
	return nil
}

// articlePage fills the template the way the Python downloader does, title without its first four characters included.
func articlePage(lecture udemy.LecturePlan) string {
	title := string([]rune(lecture.Title)[min(lectureNumberPrefixRunes, len([]rune(lecture.Title))):])
	page := strings.ReplaceAll(articleTemplate, titlePlaceholder, title)
	return strings.ReplaceAll(page, dataPlaceholder, lecture.HTMLBody)
}

func writeExternalLink(chapterDir string, asset udemy.AssetPlan) error {
	if err := writeLibraryFile(filepath.Join(chapterDir, asset.File), fmt.Appendf(nil, internetShortcutFormat, asset.URL)); err != nil {
		return fmt.Errorf("write link %s: %w", asset.File, err)
	}
	listPath := filepath.Join(chapterDir, externalLinksFile)
	listed, err := os.ReadFile(listPath) //nolint:gosec // reason: the list lives in the chapter directory this download writes
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("read %s: %w", externalLinksFile, err)
	}
	if slices.Contains(listedLines(listed), strings.ToLower(asset.File)) {
		return nil
	}
	list, err := os.OpenFile(listPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, libraryFileMode) //nolint:gosec // reason: a library file in the chapter directory this download writes
	if err != nil {
		return fmt.Errorf("open %s: %w", externalLinksFile, err)
	}
	_, writeErr := fmt.Fprintf(list, externalLinkEntryFormat, asset.File, asset.URL)
	if err := errors.Join(writeErr, list.Close()); err != nil {
		return fmt.Errorf("append to %s: %w", externalLinksFile, err)
	}
	return nil
}

func listedLines(listed []byte) []string {
	var lines []string
	for _, line := range lineTerminator.Split(strings.ToValidUTF8(string(listed), ""), -1) {
		lines = append(lines, strings.ToLower(strings.TrimSpace(line)))
	}
	return lines
}
