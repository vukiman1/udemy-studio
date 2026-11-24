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

// An overloaded provider usually recovers within seconds, so a model is asked again after these pauses before the next one.
var retryPauses = []time.Duration{2 * time.Second, 5 * time.Second}

// The check asks for the shortest useful answer, so it costs as little as the provider allows.
var checkPrompt = Prompt{System: "Reply with the single word OK.", User: "ping"}

// Check is what checking a key found; Model is empty when no model was asked.
type Check struct {
	Model   string
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
	Models   []string
}

// Draft is what the settings form sends; an empty APIKey keeps the saved key of the same provider.
type Draft struct {
	Provider Provider
	APIKey   string
	Models   []string
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
			view.Providers = append(view.Providers, ProviderView{Provider: profile.Provider, KeyHint: keyHint(saved.APIKey), Models: saved.Models})
		}
	}
	return view, nil
}

// Save makes the draft's provider the one in use; a new key is kept only once the provider has accepted it.
func (service *Service) Save(ctx context.Context, draft Draft) error {
	profile, err := service.profile(draft.Provider)
	if err != nil {
		return err
	}
	models := cleanModels(draft.Models)
	if len(models) == 0 {
		return ErrNoModels
	}
	key, isNew, err := service.key(ctx, draft.Provider, draft.APIKey)
	if err != nil {
		return err
	}
	if isNew {
		if err := profile.Client.CheckKey(ctx, key); err != nil {
			return fromProvider(err)
		}
	}
	config, err := service.load(ctx)
	if err != nil {
		return err
	}
	config.Providers[draft.Provider] = ProviderSettings{APIKey: key, Models: models}
	config.Active, config.IsEnabled = draft.Provider, true
	return service.store.Save(ctx, config)
}

// Check confirms the key works and, given a model, that the model answers with it, timing the answer;
// it uses what is on the form, saved or not.
func (service *Service) Check(ctx context.Context, provider Provider, typedKey, model string) (Check, error) {
	profile, err := service.profile(provider)
	if err != nil {
		return Check{}, err
	}
	key, _, err := service.key(ctx, provider, typedKey)
	if err != nil {
		return Check{}, err
	}
	if model == "" {
		return Check{}, fromProvider(profile.Client.CheckKey(ctx, key))
	}
	started := service.now()
	completion, err := profile.Client.Complete(ctx, key, model, checkPrompt)
	if err != nil {
		return Check{}, fromProvider(err)
	}
	return Check{Model: model, Latency: service.now().Sub(started), Reply: truncate(strings.TrimSpace(completion.Text), maxCheckReply)}, nil
}

func (service *Service) ListModels(ctx context.Context, provider Provider, typedKey string) ([]ModelInfo, error) {
	profile, err := service.profile(provider)
	if err != nil {
		return nil, err
	}
	key, _, err := service.key(ctx, provider, typedKey)
	if err != nil {
		return nil, err
	}
	models, err := profile.Client.ListModels(ctx, key)
	return models, fromProvider(err)
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

// Forget removes one provider's key and models; forgetting the provider in use leaves none in use.
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

// Complete asks the configured models in order and returns the first answer; a rejected key stops at once, since every model would reject it.
func (service *Service) Complete(ctx context.Context, prompt Prompt) (Completion, error) {
	settings, profile, err := service.active(ctx)
	if err != nil {
		return Completion{}, err
	}
	var failures []error
	// Failed attempts are billed too, so their tokens and cost stay in what is reported.
	var spent Completion
	for _, model := range settings.Models {
		completion, err := service.completeWithRetries(ctx, profile, settings.APIKey, model, prompt)
		spent.Usage = spent.Usage.Plus(completion.Usage)
		spent.CostUSD += profile.cost(model, completion.Usage)
		if err == nil {
			completion.Model, completion.Usage, completion.CostUSD = model, spent.Usage, spent.CostUSD
			return completion, nil
		}
		if errors.Is(err, ErrInvalidAPIKey) || ctx.Err() != nil {
			return spent, err
		}
		service.logger.Warn("model failed, trying the next one", "provider", profile.Provider, "model", model, "err", err)
		failures = append(failures, fmt.Errorf("%s: %w", model, err))
	}
	return spent, fmt.Errorf("every configured model failed: %w", errors.Join(failures...))
}

func (service *Service) completeWithRetries(ctx context.Context, profile Profile, key secret.Secret, model string, prompt Prompt) (Completion, error) {
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

// EstimateUSD prices usage at the first model, which answers unless it fails; false when that price is unknown.
func (service *Service) EstimateUSD(ctx context.Context, usage Usage) (float64, bool, error) {
	settings, profile, err := service.active(ctx)
	if err != nil {
		return 0, false, err
	}
	price, known := profile.Prices[settings.Models[0]]
	if !known {
		return 0, false, nil
	}
	return price.cost(usage), true, nil
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
	if !found || !config.IsEnabled || len(settings.Models) == 0 {
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
	return config, nil
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

func (profile Profile) cost(model string, usage Usage) float64 {
	return profile.Prices[model].cost(usage)
}

func (price Price) cost(usage Usage) float64 {
	return (float64(usage.InputTokens)*price.InputUSDPerMillion + float64(usage.OutputTokens)*price.OutputUSDPerMillion) / tokensPerMillion
}

func cleanModels(models []string) []string {
	seen := map[string]bool{}
	var cleaned []string
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model != "" && !seen[model] {
			seen[model] = true
			cleaned = append(cleaned, model)
		}
	}
	return cleaned
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
