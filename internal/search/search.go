// Package search finds the moment a word is said in any downloaded lecture, through an index of its subtitle cues.
package search

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/subtitle"
)

const (
	maxHits            = 50
	highlightStartRune = ''
	highlightEndRune   = ''
	// The snippet marks each match with these private-use characters, which never appear in subtitles.
	HighlightStart = string(highlightStartRune)
	HighlightEnd   = string(highlightEndRune)
)

// FTS5 strips combining marks but keeps đ as its own letter, so it is folded here, one rune for one rune.
var foldUnmarkedLetters = strings.NewReplacer("đ", "d", "Đ", "D")

type Entry struct {
	LectureKey string
	Language   string
	Start      time.Duration
	Text       string
	SearchText string
}

func NewEntry(lectureKey, language string, start time.Duration, text string) Entry {
	return Entry{LectureKey: lectureKey, Language: language, Start: start, Text: text, SearchText: foldUnmarkedLetters.Replace(text)}
}

// Match is an indexed entry whose SearchText has each match between HighlightStart and HighlightEnd.
type Match struct {
	Entry
	HighlightedSearchText string
}

type Hit struct {
	LectureKey string
	Language   string
	Start      time.Duration
	Snippet    string
}

type Index interface {
	Replace(ctx context.Context, entries []Entry) error
	// Search matches entries whose SearchText holds every term, the last one as a prefix so results follow the typing.
	Search(ctx context.Context, terms []string, limit int) ([]Match, error)
}

type Service struct {
	index  Index
	logger *slog.Logger
	// rebuilding keeps an older library snapshot from committing after a newer one.
	rebuilding sync.Mutex
}

func NewService(index Index, logger *slog.Logger) *Service {
	return &Service{index: index, logger: logger}
}

// Rebuild indexes every subtitle cue of the library; an unreadable file is skipped as the player skips it.
func (service *Service) Rebuild(ctx context.Context, catalog *library.Catalog, media fs.FS) error {
	service.rebuilding.Lock()
	defer service.rebuilding.Unlock()
	var entries []Entry
	for _, course := range catalog.Current().Courses {
		for _, chapter := range course.Chapters {
			for _, lecture := range chapter.Lectures {
				entries = append(entries, service.lectureEntries(lecture, media)...)
			}
		}
	}
	if err := service.index.Replace(ctx, entries); err != nil {
		return fmt.Errorf("index %d subtitle cues: %w", len(entries), err)
	}
	return nil
}

func (service *Service) lectureEntries(lecture library.Lecture, media fs.FS) []Entry {
	var entries []Entry
	for _, file := range lecture.Subtitles {
		cues, err := subtitle.ParseSRTFile(media, file.Path)
		if err != nil {
			service.logger.Warn("not indexing unreadable subtitle", "path", file.Path, "err", err)
			continue
		}
		for _, cue := range cues {
			entries = append(entries, NewEntry(lecture.Key, file.Language, cue.Start, cue.Text))
		}
	}
	return entries
}

func (service *Service) Search(ctx context.Context, query string) ([]Hit, error) {
	terms := strings.Fields(foldUnmarkedLetters.Replace(query))
	if len(terms) == 0 {
		return nil, nil
	}
	matches, err := service.index.Search(ctx, terms, maxHits)
	if err != nil {
		return nil, fmt.Errorf("search subtitles for %q: %w", query, err)
	}
	hits := make([]Hit, len(matches))
	for position, match := range matches {
		hits[position] = Hit{
			LectureKey: match.LectureKey,
			Language:   match.Language,
			Start:      match.Start,
			Snippet:    highlightOriginal(match.HighlightedSearchText, match.Text),
		}
	}
	return hits, nil
}

// highlightOriginal moves the markers onto the original text, which folding kept rune for rune.
func highlightOriginal(highlighted, original string) string {
	originalRunes := []rune(original)
	var restored strings.Builder
	position := 0
	for _, char := range highlighted {
		if char == highlightStartRune || char == highlightEndRune || position >= len(originalRunes) {
			restored.WriteRune(char)
			continue
		}
		restored.WriteRune(originalRunes[position])
		position++
	}
	return restored.String()
}
