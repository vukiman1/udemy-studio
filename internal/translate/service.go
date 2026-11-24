package translate

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"strings"

	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/subtitle"
)

const (
	cuesPerRequest   = 120
	attemptsPerChunk = 3
	// Weaker models stop a few cues short of a long list unless told where it ends.
	userPromptFormat = "Lecture: %s\n\n%sTranslate these %d subtitle cues, %d to %d, into Vietnamese. Reply with all %d lines.\n\n%s"
	// A chunk starts mid-conversation, so it sees how the last few cues before it were translated.
	contextCues   = 5
	contextHeader = "The cues just before these, already translated, for context only. Do not translate or repeat them:\n"
	contextFormat = "#%d %s → %s\n"
	driftReminder = "\nAn earlier attempt slipped lines into the neighbouring cue. Keep each cue's own meaning in that cue.\n"
)

//go:embed system_prompt.txt
var systemPrompt string

type Service struct {
	ai     AI
	logger *slog.Logger
}

func NewService(ai AI, logger *slog.Logger) *Service {
	return &Service{ai: ai, logger: logger}
}

// TranslateCues returns the cues with Vietnamese text and the original timings, once every chunk parses and no span drifts.
func (service *Service) TranslateCues(ctx context.Context, lectureTitle string, cues []subtitle.Cue) ([]subtitle.Cue, Spend, error) {
	chunks := chunkCues(cues)
	translations := make([]map[int]string, len(chunks))
	var spend Spend
	for position, chunk := range chunks {
		translated, used, err := service.translateChunk(ctx, lectureTitle, chunk, contextBefore(chunks, translations, position), false)
		spend = Spend{Usage: spend.Usage.Plus(used.Usage), CostUSD: spend.CostUSD + used.CostUSD}
		if err != nil {
			return nil, spend, err
		}
		translations[position] = translated
	}
	retried := map[int]bool{}
	for {
		result := applyTranslations(cues, translations)
		drifted := chunksTouching(driftSpans(texts(cues), texts(result)))
		if len(drifted) == 0 {
			return result, spend, nil
		}
		for _, position := range drifted {
			if retried[position] {
				return nil, spend, fmt.Errorf("%w: cues %d-%d of %q", ErrDrift, chunks[position][0].Index, chunks[position][len(chunks[position])-1].Index, lectureTitle)
			}
			retried[position] = true
			translated, used, err := service.translateChunk(ctx, lectureTitle, chunks[position], contextBefore(chunks, translations, position), true)
			spend = Spend{Usage: spend.Usage.Plus(used.Usage), CostUSD: spend.CostUSD + used.CostUSD}
			if err != nil {
				return nil, spend, err
			}
			translations[position] = translated
		}
	}
}

func (service *Service) translateChunk(ctx context.Context, lectureTitle string, chunk []subtitle.Cue, earlier string, afterDrift bool) (map[int]string, Spend, error) {
	lines := cueLines(chunk)
	if afterDrift {
		lines += driftReminder
	}
	first, last := chunk[0].Index, chunk[len(chunk)-1].Index
	prompt := llm.Prompt{System: systemPrompt, User: fmt.Sprintf(userPromptFormat, lectureTitle, earlier, len(chunk), first, last, len(chunk), lines)}
	var spend Spend
	var lastErr error
	for attempt := 1; attempt <= attemptsPerChunk; attempt++ {
		completion, err := service.ai.Complete(ctx, prompt)
		spend = spend.plus(completion)
		if err != nil {
			return nil, spend, fmt.Errorf("translate cues %d-%d of %q: %w", chunk[0].Index, chunk[len(chunk)-1].Index, lectureTitle, err)
		}
		translated, err := parseReply(completion.Text, chunk)
		if err == nil {
			return translated, spend, nil
		}
		lastErr = err
		service.logger.Warn("asking again for a reply that does not match its cues", "lecture", lectureTitle, "attempt", attempt, "err", err)
	}
	return nil, spend, fmt.Errorf("cues %d-%d of %q after %d attempts: %w", chunk[0].Index, chunk[len(chunk)-1].Index, lectureTitle, attemptsPerChunk, lastErr)
}

// contextBefore shows the end of the previous chunk next to its translation, without tabs, so it never reads as cues to translate.
func contextBefore(chunks [][]subtitle.Cue, translations []map[int]string, position int) string {
	if position == 0 || translations[position-1] == nil {
		return ""
	}
	previous := chunks[position-1]
	var block strings.Builder
	block.WriteString(contextHeader)
	for _, cue := range previous[max(0, len(previous)-contextCues):] {
		fmt.Fprintf(&block, contextFormat, cue.Index, singleLine(cue.Text), translations[position-1][cue.Index])
	}
	return block.String() + "\n"
}

func chunkCues(cues []subtitle.Cue) [][]subtitle.Cue {
	var chunks [][]subtitle.Cue
	for start := 0; start < len(cues); start += cuesPerRequest {
		chunks = append(chunks, cues[start:min(start+cuesPerRequest, len(cues))])
	}
	return chunks
}

func chunksTouching(spans []span) []int {
	touched := map[int]bool{}
	var positions []int
	for _, drift := range spans {
		for chunk := drift.from / cuesPerRequest; chunk <= (drift.to-1)/cuesPerRequest; chunk++ {
			if !touched[chunk] {
				touched[chunk] = true
				positions = append(positions, chunk)
			}
		}
	}
	return positions
}

func applyTranslations(cues []subtitle.Cue, translations []map[int]string) []subtitle.Cue {
	result := make([]subtitle.Cue, len(cues))
	for position, cue := range cues {
		cue.Text = translations[position/cuesPerRequest][cue.Index]
		result[position] = cue
	}
	return result
}

func texts(cues []subtitle.Cue) []string {
	all := make([]string, len(cues))
	for position, cue := range cues {
		all[position] = cue.Text
	}
	return all
}
