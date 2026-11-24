// Package claude answers prompts with the Claude API through Anthropic's Go SDK.
package claude

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"

	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

// A chunk of 120 cues replies in well under this; the rest is room for thinking.
const maxReplyTokens = 16000

var (
	// errForbidden is a key that may not use one model: a rejected key when checking it, but another model may still answer.
	errForbidden = errors.New("the key may not use this")
	ErrRefused   = errors.New("claude declined to answer")
	ErrTruncated = errors.New("claude's reply was cut off before it ended")
)

// Server-side fallback re-serves a declined request with another model inside the same call; these models offer it.
var modelsWithServerFallback = map[string]bool{"claude-opus-5-5": true, "claude-sonnet-5-5": true}

type Client struct{}

func Profile() llm.Profile {
	return llm.Profile{
		Provider: llm.ProviderClaude,
		Label:    "Claude (Anthropic)",
		Known: []llm.ModelInfo{
			{ID: "claude-opus-5-5", Label: "Claude Opus 5.5"},
			{ID: "claude-sonnet-5-5", Label: "Claude Sonnet 5.5"},
			{ID: "claude-haiku-5-5", Label: "Claude Haiku 5.5"},
		},
		// Measured on a 303-cue lecture: Haiku kept the cues and terms at 1/25 of Opus's cost; Opus stands by.
		// Sonnet is left out of the chain: it merged lines and drifted 25 cues.
		Recommended: []string{"claude-haiku-5-5", "claude-opus-5-5"},
		Prices: map[string]llm.Price{
			"claude-opus-5-5":   {InputUSDPerMillion: 4, OutputUSDPerMillion: 20},
			"claude-sonnet-5-5": {InputUSDPerMillion: 2, OutputUSDPerMillion: 10},
			"claude-haiku-5-5":  {InputUSDPerMillion: 0.10, OutputUSDPerMillion: 0.50},
		},
		Client: Client{},
	}
}

func (Client) CheckKey(ctx context.Context, key secret.Secret) error {
	_, err := Client{}.ListModels(ctx, key)
	if errors.Is(err, errForbidden) {
		return fmt.Errorf("%w: %w", llm.ErrInvalidAPIKey, err)
	}
	return err
}

func (Client) ListModels(ctx context.Context, key secret.Secret) ([]llm.ModelInfo, error) {
	client := anthropic.NewClient(option.WithAPIKey(key.Reveal()))
	pages := client.Models.ListAutoPaging(ctx, anthropic.ModelListParams{})
	var models []llm.ModelInfo
	for pages.Next() {
		model := pages.Current()
		models = append(models, llm.ModelInfo{ID: model.ID, Label: model.DisplayName})
	}
	if err := pages.Err(); err != nil {
		return nil, fmt.Errorf("list Claude models: %w", classify(err))
	}
	return models, nil
}

func (Client) Complete(ctx context.Context, key secret.Secret, model string, prompt llm.Prompt) (llm.Completion, error) {
	client := anthropic.NewClient(option.WithAPIKey(key.Reveal()))
	params := anthropic.BetaMessageNewParams{
		Model:        model,
		MaxTokens:    maxReplyTokens,
		System:       []anthropic.BetaTextBlockParam{{Text: prompt.System, CacheControl: anthropic.NewBetaCacheControlEphemeralParam()}},
		Messages:     []anthropic.BetaMessageParam{anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(prompt.User))},
		OutputConfig: anthropic.BetaOutputConfigParam{Effort: anthropic.BetaOutputConfigEffortMedium},
	}
	if modelsWithServerFallback[model] {
		params.Fallbacks = anthropic.BetaFallbacksParamOfDefault()
		params.Betas = []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01}
	}
	message, err := client.Beta.Messages.New(ctx, params)
	if err != nil {
		return llm.Completion{}, fmt.Errorf("call %s: %w", model, classify(err))
	}
	usage := llm.Usage{
		InputTokens:  message.Usage.InputTokens + message.Usage.CacheReadInputTokens + message.Usage.CacheCreationInputTokens,
		OutputTokens: message.Usage.OutputTokens,
	}
	if message.StopReason == anthropic.BetaStopReasonRefusal {
		return llm.Completion{Usage: usage}, fmt.Errorf("%w (%s)", ErrRefused, message.StopDetails.Category)
	}
	if message.StopReason == anthropic.BetaStopReasonMaxTokens || message.StopReason == anthropic.BetaStopReasonModelContextWindowExceeded {
		return llm.Completion{Usage: usage}, fmt.Errorf("%w (%s)", ErrTruncated, message.StopReason)
	}
	var text strings.Builder
	for _, block := range message.Content {
		if textBlock, ok := block.AsAny().(anthropic.BetaTextBlock); ok {
			text.WriteString(textBlock.Text)
		}
	}
	return llm.Completion{Text: text.String(), Usage: usage}, nil
}

// classify marks a rejected key, which stops a run, and an overloaded API, which is worth asking again.
func classify(err error) error {
	var apiErr *anthropic.Error
	if !errors.As(err, &apiErr) {
		return err
	}
	switch {
	case apiErr.StatusCode == http.StatusUnauthorized:
		return fmt.Errorf("%w by Claude (HTTP %d)", llm.ErrInvalidAPIKey, apiErr.StatusCode)
	case apiErr.StatusCode == http.StatusForbidden:
		return fmt.Errorf("%w (HTTP %d): %w", errForbidden, apiErr.StatusCode, err)
	case apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= http.StatusInternalServerError:
		// The SDK has already retried this twice; marking it lets the next attempt wait longer.
		return fmt.Errorf("%w: %w", llm.ErrTemporary, err)
	}
	return err
}
