package llm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	workingKey = "sk-test-working-1234" //nolint:gosec // reason: fake key for the fake provider
	primary    = "big-model"
	standby    = "small-model"
)

// fakeClient accepts only workingKey and fails the models listed in failing.
type fakeClient struct {
	failing map[string]error
	// billsFailures makes a failure still cost what was produced, as a truncated answer does.
	billsFailures bool
	asked         []string
}

func (client *fakeClient) CheckKey(_ context.Context, key secret.Secret) error {
	if key.Reveal() != workingKey {
		return fmt.Errorf("%w by the fake provider", ErrInvalidAPIKey)
	}
	return nil
}

func (client *fakeClient) ListModels(context.Context, secret.Secret) ([]ModelInfo, error) {
	return nil, nil
}

func (client *fakeClient) Complete(ctx context.Context, key secret.Secret, model string, _ Prompt) (Completion, error) {
	client.asked = append(client.asked, model)
	if err := client.CheckKey(ctx, key); err != nil {
		return Completion{}, err
	}
	if err := client.failing[model]; err != nil {
		if client.billsFailures {
			return Completion{Usage: Usage{InputTokens: 1_000_000, OutputTokens: 1_000_000}}, err
		}
		return Completion{}, err
	}
	return Completion{Text: "answer from " + model, Usage: Usage{InputTokens: 1_000_000, OutputTokens: 1_000_000}}, nil
}

type memoryStore struct {
	config Config
	found  bool
}

func (store *memoryStore) Load(context.Context) (Config, bool, error) {
	return store.config, store.found, nil
}

func (store *memoryStore) Save(_ context.Context, config Config) error {
	store.config, store.found = config, true
	return nil
}

func (store *memoryStore) Delete(context.Context) error {
	store.config, store.found = Config{}, false
	return nil
}

func newTestService(client *fakeClient, store *memoryStore) *Service {
	profiles := []Profile{
		{Provider: ProviderClaude, Client: client, Prices: map[string]Price{primary: {InputUSDPerMillion: 4, OutputUSDPerMillion: 20}, standby: {InputUSDPerMillion: 1, OutputUSDPerMillion: 5}}},
		{Provider: ProviderOpenAI, Client: client},
	}
	return NewService(store, profiles, time.Now, slog.New(slog.DiscardHandler))
}

func savedService(t *testing.T, client *fakeClient) *Service {
	t.Helper()
	service := newTestService(client, &memoryStore{})
	if err := service.Save(t.Context(), Draft{Provider: ProviderClaude, APIKey: workingKey, Models: []string{primary, standby}}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return service
}

func TestAFailingModelFallsBackToTheNextAndIsPricedByTheOneThatAnswered(t *testing.T) {
	client := &fakeClient{failing: map[string]error{primary: errors.New("overloaded")}}
	completion, err := savedService(t, client).Complete(t.Context(), Prompt{})
	if err != nil || completion.Model != standby || completion.CostUSD != 6 {
		t.Errorf("Complete = %+v, %v; want the standby's answer priced at $1+$5", completion, err)
	}
}

func TestARejectedKeyStopsWithoutTryingTheOtherModels(t *testing.T) {
	store := &memoryStore{}
	client := &fakeClient{}
	service := newTestService(client, store)
	store.config, store.found = Config{
		Active:    ProviderClaude,
		Providers: map[Provider]ProviderSettings{ProviderClaude: {APIKey: secret.New("sk-revoked"), Models: []string{primary, standby}}},
		IsEnabled: true,
	}, true
	if _, err := service.Complete(t.Context(), Prompt{}); !errors.Is(err, ErrInvalidAPIKey) || len(client.asked) != 1 {
		t.Errorf("err = %v after asking %v; want ErrInvalidAPIKey after one model", err, client.asked)
	}
}

func TestSavingKeepsAWorkingKeyUntilANewOneIsAccepted(t *testing.T) {
	service := savedService(t, &fakeClient{})
	if err := service.Save(t.Context(), Draft{Provider: ProviderClaude, APIKey: "sk-typo", Models: []string{standby}}); !errors.Is(err, ErrInvalidAPIKey) {
		t.Errorf("Save with a rejected key: %v, want ErrInvalidAPIKey", err)
	}
	if err := service.Save(t.Context(), Draft{Provider: ProviderClaude, Models: []string{standby}}); err != nil {
		t.Errorf("Save with the key left empty: %v, want the saved key reused", err)
	}
	if err := service.Save(t.Context(), Draft{Provider: ProviderOpenAI, Models: []string{"gpt"}}); !errors.Is(err, ErrNoKey) {
		t.Errorf("Save for another provider without a key: %v, want ErrNoKey", err)
	}
	view, err := service.View(t.Context())
	if err != nil || view.Active != ProviderClaude || len(view.Providers) != 1 {
		t.Fatalf("view = %+v, %v; want only Claude saved and in use", view, err)
	}
	if saved := view.Providers[0]; len(saved.Models) != 1 || strings.Contains(saved.KeyHint, "working") || !strings.HasSuffix(saved.KeyHint, "1234") {
		t.Errorf("saved Claude = %+v; want the first save's key, masked", saved)
	}
}

func TestAStoppedConfigurationAnswersNothing(t *testing.T) {
	service := savedService(t, &fakeClient{})
	if err := service.SetEnabled(t.Context(), false); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}
	if _, err := service.Complete(t.Context(), Prompt{}); !errors.Is(err, ErrNotConfigured) {
		t.Errorf("Complete while stopped: %v, want ErrNotConfigured", err)
	}
}

func TestCheckingWithAModelAsksThatModelAndOnlyChecksTheKeyWithout(t *testing.T) {
	client := &fakeClient{}
	service := savedService(t, client)
	check, err := service.Check(t.Context(), ProviderClaude, "", standby)
	if err != nil || check.Model != standby || check.Reply != "answer from "+standby || len(client.asked) != 1 {
		t.Errorf("Check with a model = %+v, %v after asking %v", check, err, client.asked)
	}
	if check, err := service.Check(t.Context(), ProviderClaude, "", ""); err != nil || check.Model != "" || len(client.asked) != 1 {
		t.Errorf("Check without a model = %+v, %v after asking %v; want the key checked only", check, err, client.asked)
	}
}

func TestABusyModelIsAskedAgainBeforeTheNextOne(t *testing.T) {
	client := &fakeClient{failing: map[string]error{primary: fmt.Errorf("%w: overloaded", ErrTemporary)}}
	service := savedService(t, client)
	service.pause = func(context.Context, time.Duration) error { return nil }
	completion, err := service.Complete(t.Context(), Prompt{})
	wantAsked := []string{primary, primary, primary, standby}
	if err != nil || completion.Model != standby || strings.Join(client.asked, ",") != strings.Join(wantAsked, ",") {
		t.Errorf("Complete = %+v, %v after asking %v; want %v", completion, err, client.asked, wantAsked)
	}
}

func TestEachProviderKeepsItsOwnKeyWhenAnotherIsInUse(t *testing.T) {
	service := savedService(t, &fakeClient{})
	if err := service.Save(t.Context(), Draft{Provider: ProviderOpenAI, APIKey: workingKey, Models: []string{"gpt"}}); err != nil {
		t.Fatalf("Save OpenAI: %v", err)
	}
	if err := service.Save(t.Context(), Draft{Provider: ProviderClaude, Models: []string{primary}}); err != nil {
		t.Errorf("switching back to Claude without typing its key again: %v", err)
	}
	if err := service.Forget(t.Context(), ProviderOpenAI); err != nil {
		t.Fatalf("Forget OpenAI: %v", err)
	}
	view, err := service.View(t.Context())
	if err != nil || view.Active != ProviderClaude || len(view.Providers) != 1 || view.Providers[0].Provider != ProviderClaude {
		t.Errorf("view = %+v, %v; want Claude still saved and in use", view, err)
	}
}

// unreadableStore fails to open what it holds, as after the sealing key file was lost.
type unreadableStore struct{ memoryStore }

func (store *unreadableStore) Load(ctx context.Context) (Config, bool, error) {
	if !store.found {
		return Config{}, true, fmt.Errorf("%w: sealed with another key", ErrUnreadableConfig)
	}
	return store.memoryStore.Load(ctx)
}

func TestSettingsThatCannotBeOpenedCanBeSavedOver(t *testing.T) {
	store := &unreadableStore{}
	service := NewService(store, []Profile{{Provider: ProviderClaude, Client: &fakeClient{}}}, time.Now, slog.New(slog.DiscardHandler))
	if view, err := service.View(t.Context()); err != nil || view.Active != "" {
		t.Fatalf("View of unreadable settings = %+v, %v; want empty settings", view, err)
	}
	if err := service.Save(t.Context(), Draft{Provider: ProviderClaude, APIKey: workingKey, Models: []string{primary}}); err != nil {
		t.Errorf("Save over unreadable settings: %v", err)
	}
}

func TestATruncatedAnswerStillCountsTowardsTheCost(t *testing.T) {
	client := &fakeClient{failing: map[string]error{primary: errors.New("cut off at max tokens")}, billsFailures: true}
	service := savedService(t, client)
	completion, err := service.Complete(t.Context(), Prompt{})
	if err != nil || completion.CostUSD != 24+6 || completion.Usage.OutputTokens != 2_000_000 {
		t.Errorf("Complete = %+v, %v; want the failed big model's $24 added to the standby's $6", completion, err)
	}
}
