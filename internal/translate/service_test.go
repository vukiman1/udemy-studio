package translate

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/subtitle"
)

const translatedPrefix = "VI "

// scriptedModel answers every prompt by rewriting each "<index>\t<text>" line through reply.
type scriptedModel struct {
	mutex   sync.Mutex
	calls   int
	prompts []string
	reply   func(call int, lines [][2]string) string
}

func (model *scriptedModel) Complete(_ context.Context, prompt llm.Prompt) (llm.Completion, error) {
	model.mutex.Lock()
	model.calls++
	model.prompts = append(model.prompts, prompt.User)
	call := model.calls
	model.mutex.Unlock()
	var lines [][2]string
	for _, line := range strings.Split(prompt.User, "\n") {
		if index, text, ok := strings.Cut(line, cueSeparator); ok {
			lines = append(lines, [2]string{index, text})
		}
	}
	return llm.Completion{Text: model.reply(call, lines), Usage: llm.Usage{InputTokens: 10, OutputTokens: 20}, CostUSD: 0.01}, nil
}

func (*scriptedModel) EstimateUSD(context.Context, llm.Usage) (float64, bool, error) {
	return 1, true, nil
}

func (*scriptedModel) IsReady(context.Context) error { return nil }

func faithful(_ int, lines [][2]string) string {
	var reply strings.Builder
	for _, line := range lines {
		fmt.Fprintf(&reply, "%s\t%s%s\n", line[0], translatedPrefix, line[1])
	}
	return reply.String()
}

// shifted gives every cue its neighbour's text, the way a model merging two short cues slips a whole span.
func shifted(_ int, lines [][2]string) string {
	var reply strings.Builder
	for position, line := range lines {
		fmt.Fprintf(&reply, "%s\t%s%s\n", line[0], translatedPrefix, lines[min(position+1, len(lines)-1)][1])
	}
	return reply.String()
}

func lectureCues(count int) []subtitle.Cue {
	cues := make([]subtitle.Cue, count)
	for position := range cues {
		start := time.Duration(position) * 2 * time.Second
		cues[position] = subtitle.Cue{Index: position + 1, Start: start, End: start + time.Second, Text: strings.Repeat("word ", 1+position*7%11)}
	}
	return cues
}

func newTestService(model AI) *Service {
	return NewService(model, slog.New(slog.DiscardHandler))
}

func TestTranslationKeepsEveryTimingAndReplacesOnlyText(t *testing.T) {
	cues := lectureCues(250)
	translated, spend, err := newTestService(&scriptedModel{reply: faithful}).TranslateCues(t.Context(), "HTTP2", cues)
	if err != nil {
		t.Fatalf("TranslateCues: %v", err)
	}
	for position, cue := range translated {
		original := cues[position]
		if cue.Index != original.Index || cue.Start != original.Start || cue.End != original.End || cue.Text != translatedPrefix+singleLine(original.Text) {
			t.Fatalf("cue %d = %+v, want the timing of %+v with translated text", position, cue, original)
		}
	}
	if chunks := 3; spend.Usage.InputTokens != int64(chunks*10) {
		t.Errorf("spend = %+v, want one request per chunk of %d cues", spend, cuesPerRequest)
	}
}

func TestAReplyThatDropsACueIsAskedForAgainAndNeverAccepted(t *testing.T) {
	dropsLastCue := func(call int, lines [][2]string) string {
		if call == 1 {
			return faithful(call, lines[:len(lines)-1])
		}
		return faithful(call, lines)
	}
	model := &scriptedModel{reply: dropsLastCue}
	if _, _, err := newTestService(model).TranslateCues(t.Context(), "HTTP2", lectureCues(20)); err != nil || model.calls != 2 {
		t.Errorf("err = %v after %d calls; want success on the second", err, model.calls)
	}
	alwaysDrops := &scriptedModel{reply: func(call int, lines [][2]string) string { return faithful(call, lines[:len(lines)-1]) }}
	if _, _, err := newTestService(alwaysDrops).TranslateCues(t.Context(), "HTTP2", lectureCues(20)); !errors.Is(err, ErrMalformedReply) || alwaysDrops.calls != attemptsPerChunk {
		t.Errorf("err = %v after %d calls; want ErrMalformedReply after %d", err, alwaysDrops.calls, attemptsPerChunk)
	}
}

func TestAShiftedTranslationIsRedoneAndRejectedIfItStaysShifted(t *testing.T) {
	recovers := &scriptedModel{reply: func(call int, lines [][2]string) string {
		if call == 1 {
			return shifted(call, lines)
		}
		return faithful(call, lines)
	}}
	if _, _, err := newTestService(recovers).TranslateCues(t.Context(), "WebRTC", lectureCues(60)); err != nil || recovers.calls != 2 {
		t.Errorf("err = %v after %d calls; want the drifted chunk redone once", err, recovers.calls)
	}
	stays := &scriptedModel{reply: shifted}
	if _, _, err := newTestService(stays).TranslateCues(t.Context(), "WebRTC", lectureCues(60)); !errors.Is(err, ErrDrift) {
		t.Errorf("err = %v, want ErrDrift", err)
	}
}

func TestEachChunkSeesHowTheCuesBeforeItWereTranslated(t *testing.T) {
	cues := lectureCues(cuesPerRequest + 10)
	model := &scriptedModel{reply: faithful}
	if _, _, err := newTestService(model).TranslateCues(t.Context(), "HTTP2", cues); err != nil {
		t.Fatalf("TranslateCues: %v", err)
	}
	last := cues[cuesPerRequest-1]
	want := fmt.Sprintf("#%d %s → %s%s", last.Index, singleLine(last.Text), translatedPrefix, singleLine(last.Text))
	if strings.Contains(model.prompts[0], contextHeader) || !strings.Contains(model.prompts[1], want) {
		t.Errorf("second prompt lacks the first chunk's last translated cue %q:\n%s", want, model.prompts[1])
	}
}

func TestBlankLinesInAReplyAreNotCues(t *testing.T) {
	cues := lectureCues(3)
	reply := "1\tmột\n\n2\thai\n   \n3\tba\n"
	if translated, err := parseReply(reply, cues); err != nil || translated[3] != "ba" {
		t.Errorf("parseReply = %v, %v; want the blank lines ignored", translated, err)
	}
}
