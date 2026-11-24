// Package translate turns a lecture's English subtitles into Vietnamese with the configured AI, and accepts a translation only when a program, not the model, has checked it.
package translate

import (
	"context"
	"errors"

	"github.com/vukiman1/udemy-studio/internal/llm"
)

var (
	ErrMalformedReply = errors.New("the model's reply does not match the cues it was given")
	ErrDrift          = errors.New("the translation looks shifted against the original cues")
)

// AI is the provider configured in the settings; *llm.Service satisfies it.
type AI interface {
	Complete(ctx context.Context, prompt llm.Prompt) (llm.Completion, error)
	EstimateUSD(ctx context.Context, usage llm.Usage) (float64, bool, error)
	IsReady(ctx context.Context) error
}

type Spend struct {
	Usage   llm.Usage
	CostUSD float64
}

func (spend Spend) plus(completion llm.Completion) Spend {
	return Spend{Usage: spend.Usage.Plus(completion.Usage), CostUSD: spend.CostUSD + completion.CostUSD}
}
