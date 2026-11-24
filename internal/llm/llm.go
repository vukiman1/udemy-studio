// Package llm keeps which AI provider, key and models the app uses, and sends prompts to them, falling back model by model.
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
	ProviderGemini Provider = "gemini"
)

var (
	ErrInvalidAPIKey   = errors.New("the API key was rejected")
	ErrNotConfigured   = errors.New("no AI provider is configured and turned on")
	ErrUnknownProvider = errors.New("unknown AI provider")
	ErrNoModels        = errors.New("choose at least one model")
	ErrNoKey           = errors.New("enter an API key for this provider")
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
	Text  string
	Usage Usage
	// Model is the one that answered, which differs from the first choice after a fallback.
	Model string
	// CostUSD is zero when the model's price is not known.
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
	// CheckKey returns an error wrapping ErrInvalidAPIKey when the provider rejects the key.
	CheckKey(ctx context.Context, key secret.Secret) error
	ListModels(ctx context.Context, key secret.Secret) ([]ModelInfo, error)
	// Complete returns an error wrapping ErrInvalidAPIKey when the provider rejects the key.
	Complete(ctx context.Context, key secret.Secret, model string, prompt Prompt) (Completion, error)
}

// Profile is what the app knows about a provider without asking it: the models it can price, best first,
// and the chain it recommends, the first model answering and the rest standing by.
type Profile struct {
	Provider    Provider
	Label       string
	Known       []ModelInfo
	Recommended []string
	Prices      map[string]Price
	Client      Client
}

// ProviderSettings is one provider's key and model chain; each provider keeps its own when another is in use.
type ProviderSettings struct {
	APIKey secret.Secret
	Models []string
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
