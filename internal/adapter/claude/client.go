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
const (
	apiURL         = "https://api.anthropic.com"
	maxReplyTokens = 16000
)

var (
	ErrRefused   = errors.New("claude declined to answer")
	ErrTruncated = errors.New("claude's reply was cut off before it ended")
)

type Client struct {
	baseURL string
}

func Profile() llm.Profile {
	return llm.Profile{
		Provider: llm.ProviderClaude,
		Label:    "Claude (Anthropic)",
		// Measured on a 303-cue lecture: Haiku kept the cues and terms at 1/25 of Opus's cost, while Sonnet drifted 25 cues.
		Model:  llm.ModelInfo{ID: "claude-haiku-5-5", Label: "Claude Haiku 5.5"},
		Price:  llm.Price{InputUSDPerMillion: 0.10, OutputUSDPerMillion: 0.50},
		Client: Client{baseURL: apiURL},
	}
}

func (claude Client) Complete(ctx context.Context, key secret.Secret, model string, prompt llm.Prompt) (llm.Completion, error) {
	client := anthropic.NewClient(option.WithAPIKey(key.Reveal()), option.WithBaseURL(claude.baseURL))
	params := anthropic.BetaMessageNewParams{
		Model:        model,
		MaxTokens:    maxReplyTokens,
		System:       []anthropic.BetaTextBlockParam{{Text: prompt.System, CacheControl: anthropic.NewBetaCacheControlEphemeralParam()}},
		Messages:     []anthropic.BetaMessageParam{anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(prompt.User))},
		OutputConfig: anthropic.BetaOutputConfigParam{Effort: anthropic.BetaOutputConfigEffortMedium},
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

// classify marks a rejected key, which stops a run, and an overloaded API, which is worth asking again; a key that
// may not use the model is as good as rejected, since the app uses no other.
func classify(err error) error {
	var apiErr *anthropic.Error
	if !errors.As(err, &apiErr) {
		return err
	}
	switch {
	case apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden:
		return fmt.Errorf("%w by Claude (HTTP %d)", llm.ErrInvalidAPIKey, apiErr.StatusCode)
	case apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= http.StatusInternalServerError:
		// The SDK has already retried this twice; marking it lets the next attempt wait longer.
		return fmt.Errorf("%w: %w", llm.ErrTemporary, err)
	}
	return err
}
