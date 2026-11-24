// Package llm keeps which AI provider and key the app uses, and sends prompts to that provider's one model.
package llm

import (
	"context"
	"errors"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

type Provider string

const (
	ProviderClaude Provider = "claude"
	ProviderOpenAI Provider = "openai"
)

var (
	ErrInvalidAPIKey   = errors.New("the API key was rejected")
	ErrNotConfigured   = errors.New("no AI provider is configured and turned on")
	ErrUnknownProvider = errors.New("unknown AI provider")
	ErrNoKey           = errors.New("enter an API key for this provider")
	// ErrQuotaExhausted means the provider will not answer again for a long while, so a run stops rather than fail lecture by lecture.
	ErrQuotaExhausted = errors.New("the provider's quota is used up")
	// ErrTemporary marks a failure worth trying the same model again for, such as an overloaded provider.
	ErrTemporary = errors.New("the provider is temporarily unavailable")
	// ErrProvider marks a failure reported by, or on the way to, the provider rather than by the app itself.
	ErrProvider = errors.New("the AI provider failed")
	// ErrUnreadableConfig means the saved settings cannot be opened, for example after the sealing key was lost.
	ErrUnreadableConfig = errors.New("the saved AI settings cannot be read")
)

type Prompt struct {
	System string
	User   string
}

type Usage struct {
	InputTokens  int64
	OutputTokens int64
}

func (usage Usage) Plus(other Usage) Usage {
	return Usage{InputTokens: usage.InputTokens + other.InputTokens, OutputTokens: usage.OutputTokens + other.OutputTokens}
}

type Completion struct {
	Text    string
	Usage   Usage
	CostUSD float64
}

type ModelInfo struct {
	ID    string
	Label string
}

type Price struct {
	InputUSDPerMillion  float64
	OutputUSDPerMillion float64
}

// Client talks to one provider; adapters implement it.
type Client interface {
	// Complete returns an error wrapping ErrInvalidAPIKey when the provider rejects the key.
	Complete(ctx context.Context, key secret.Secret, model string, prompt Prompt) (Completion, error)
}

// Profile is a provider and the one model the app uses with it, chosen by measuring translations; there is no
// fallback, so a failure never moves a run onto a pricier model.
type Profile struct {
	Provider Provider
	Label    string
	Model    ModelInfo
	Price    Price
	Client   Client
}

// ProviderSettings is one provider's key; each provider keeps its own when another is in use.
type ProviderSettings struct {
	APIKey secret.Secret
}

type Config struct {
	// Active is the provider the app uses; empty when none is chosen.
	Active    Provider
	Providers map[Provider]ProviderSettings
	IsEnabled bool
}

type ConfigStore interface {
	Load(ctx context.Context) (Config, bool, error)
	Save(ctx context.Context, config Config) error
	Delete(ctx context.Context) error
}
