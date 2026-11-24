package llm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	keyHintEdge      = 4
	keyHintMask      = "••••••••"
	tokensPerMillion = 1_000_000
	maxCheckReply    = 200
)

// An overloaded provider usually recovers within seconds, so the model is asked again after these pauses before giving up.
var retryPauses = []time.Duration{2 * time.Second, 5 * time.Second}

// The check asks for the shortest useful answer, so it costs as little as the provider allows.
var checkPrompt = Prompt{System: "Reply with the single word OK.", User: "ping"}

type Check struct {
	Latency time.Duration
	Reply   string
}

// View describes the saved configuration without any key, which never leaves the server.
type View struct {
	Active    Provider
	IsEnabled bool
	Providers []ProviderView
}

type ProviderView struct {
	Provider Provider
	KeyHint  string
}

// Draft is what the settings form sends; an empty APIKey keeps the saved key of the same provider.
type Draft struct {
	Provider Provider
	APIKey   string
}

type Service struct {
	store    ConfigStore
	profiles []Profile
	now      func() time.Time
	pause    func(ctx context.Context, duration time.Duration) error
	logger   *slog.Logger
}

func NewService(store ConfigStore, profiles []Profile, now func() time.Time, logger *slog.Logger) *Service {
	return &Service{store: store, profiles: profiles, now: now, pause: sleep, logger: logger}
}

func sleep(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (service *Service) Profiles() []Profile { return service.profiles }

func (service *Service) View(ctx context.Context) (View, error) {
	config, err := service.load(ctx)
	if err != nil {
		return View{}, err
	}
	view := View{Active: config.Active, IsEnabled: config.IsEnabled}
	for _, profile := range service.profiles {
		if saved, found := config.Providers[profile.Provider]; found {
			view.Providers = append(view.Providers, ProviderView{Provider: profile.Provider, KeyHint: keyHint(saved.APIKey)})
		}
	}
	return view, nil
}

// Save makes the draft's provider the one in use; a new key is kept only once the provider's model has answered with it,
// since a key may list models yet not be allowed this one.
func (service *Service) Save(ctx context.Context, draft Draft) error {
	profile, err := service.profile(draft.Provider)
	if err != nil {
		return err
	}
	key, isNew, err := service.key(ctx, draft.Provider, draft.APIKey)
	if err != nil {
		return err
	}
	if isNew {
		if _, err := service.completeWithRetries(ctx, profile, key, checkPrompt); err != nil {
			return fromProvider(err)
		}
	}
	config, err := service.load(ctx)
	if err != nil {
		return err
	}
	config.Providers[draft.Provider] = ProviderSettings{APIKey: key}
	config.Active, config.IsEnabled = draft.Provider, true
	return service.store.Save(ctx, config)
}

// Check asks the provider's model a short question with the key on the form, saved or not, and times the answer.
func (service *Service) Check(ctx context.Context, provider Provider, typedKey string) (Check, error) {
	profile, err := service.profile(provider)
	if err != nil {
		return Check{}, err
	}
	key, _, err := service.key(ctx, provider, typedKey)
	if err != nil {
		return Check{}, err
	}
	started := service.now()
	completion, err := service.completeWithRetries(ctx, profile, key, checkPrompt)
	if err != nil {
		return Check{}, fromProvider(err)
	}
	return Check{Latency: service.now().Sub(started), Reply: truncate(strings.TrimSpace(completion.Text), maxCheckReply)}, nil
}

func (service *Service) SetEnabled(ctx context.Context, isEnabled bool) error {
	config, err := service.load(ctx)
	if err != nil {
		return err
	}
	if config.Active == "" {
		return ErrNotConfigured
	}
	config.IsEnabled = isEnabled
	return service.store.Save(ctx, config)
}

// Forget removes one provider's key; forgetting the provider in use leaves none in use.
func (service *Service) Forget(ctx context.Context, provider Provider) error {
	config, err := service.load(ctx)
	if err != nil {
		return err
	}
	delete(config.Providers, provider)
	if config.Active == provider {
		config.Active, config.IsEnabled = "", false
	}
	if len(config.Providers) == 0 {
		return service.store.Delete(ctx)
	}
	return service.store.Save(ctx, config)
}

// Complete asks the active provider's model, again while the provider is busy; failed attempts are billed too, so
// their tokens and cost stay in what is returned, error or not.
func (service *Service) Complete(ctx context.Context, prompt Prompt) (Completion, error) {
	settings, profile, err := service.active(ctx)
	if err != nil {
		return Completion{}, err
	}
	completion, err := service.completeWithRetries(ctx, profile, settings.APIKey, prompt)
	completion.CostUSD = profile.Price.cost(completion.Usage)
	return completion, err
}

func (service *Service) completeWithRetries(ctx context.Context, profile Profile, key secret.Secret, prompt Prompt) (Completion, error) {
	model := profile.Model.ID
	completion, err := profile.Client.Complete(ctx, key, model, prompt)
	usage := completion.Usage
	for _, pause := range retryPauses {
		if !errors.Is(err, ErrTemporary) {
			break
		}
		service.logger.Warn("provider busy, asking the same model again", "provider", profile.Provider, "model", model, "after", pause, "err", err)
		if err := service.pause(ctx, pause); err != nil {
			return Completion{Usage: usage}, err
		}
		completion, err = profile.Client.Complete(ctx, key, model, prompt)
		usage = usage.Plus(completion.Usage)
	}
	completion.Usage = usage
	return completion, err
}

func (service *Service) EstimateUSD(ctx context.Context, usage Usage) (float64, error) {
	_, profile, err := service.active(ctx)
	if err != nil {
		return 0, err
	}
	return profile.Price.cost(usage), nil
}

func (service *Service) IsReady(ctx context.Context) error {
	_, _, err := service.active(ctx)
	return err
}

func (service *Service) active(ctx context.Context) (ProviderSettings, Profile, error) {
	config, err := service.load(ctx)
	if err != nil {
		return ProviderSettings{}, Profile{}, err
	}
	settings, found := config.Providers[config.Active]
	if !found || !config.IsEnabled {
		return ProviderSettings{}, Profile{}, ErrNotConfigured
	}
	profile, err := service.profile(config.Active)
	return settings, profile, err
}

// load returns the saved configuration, or an empty one to fill in; settings that can no longer be opened are
// started over rather than left blocking every save.
func (service *Service) load(ctx context.Context) (Config, error) {
	config, found, err := service.store.Load(ctx)
	if errors.Is(err, ErrUnreadableConfig) {
		service.logger.Warn("starting the AI settings over: the saved ones cannot be read", "err", err)
		return Config{Providers: map[Provider]ProviderSettings{}}, nil
	}
	if err != nil {
		return Config{}, err
	}
	if !found || config.Providers == nil {
		config.Providers = map[Provider]ProviderSettings{}
	}
	service.dropRetiredProviders(&config)
	return config, nil
}

// dropRetiredProviders forgets the keys of providers the app no longer offers, such as Gemini, so none stays in use.
func (service *Service) dropRetiredProviders(config *Config) {
	for provider := range config.Providers {
		if _, err := service.profile(provider); err != nil {
			delete(config.Providers, provider)
		}
	}
	if _, found := config.Providers[config.Active]; !found {
		config.Active, config.IsEnabled = "", false
	}
}

// key returns the typed key, or the provider's saved one when nothing was typed.
func (service *Service) key(ctx context.Context, provider Provider, typed string) (secret.Secret, bool, error) {
	if trimmed := strings.TrimSpace(typed); trimmed != "" {
		return secret.New(trimmed), true, nil
	}
	config, err := service.load(ctx)
	if err != nil {
		return secret.Secret{}, false, err
	}
	saved, found := config.Providers[provider]
	if !found {
		return secret.Secret{}, false, ErrNoKey
	}
	return saved.APIKey, false, nil
}

func (service *Service) profile(provider Provider) (Profile, error) {
	for _, profile := range service.profiles {
		if profile.Provider == provider {
			return profile, nil
		}
	}
	return Profile{}, fmt.Errorf("%w: %q", ErrUnknownProvider, provider)
}

func (price Price) cost(usage Usage) float64 {
	return (float64(usage.InputTokens)*price.InputUSDPerMillion + float64(usage.OutputTokens)*price.OutputUSDPerMillion) / tokensPerMillion
}

func keyHint(key secret.Secret) string {
	value := key.Reveal()
	if utf8.RuneCountInString(value) <= 2*keyHintEdge {
		return keyHintMask
	}
	return value[:keyHintEdge] + keyHintMask + value[len(value)-keyHintEdge:]
}

func truncate(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "…"
}

// fromProvider marks an error that came from the provider, keeping a rejected key recognisable as such.
func fromProvider(err error) error {
	if err == nil || errors.Is(err, ErrInvalidAPIKey) {
		return err
	}
	return fmt.Errorf("%w: %w", ErrProvider, err)
}
