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
	workingKey  = "sk-test-working-1234" //nolint:gosec // reason: fake key for the fake provider
	claudeModel = "claude-model"
	openAIModel = "openai-model"
)

// fakeClient accepts only workingKey and fails the models listed in failing.
type fakeClient struct {
	failing map[string]error
	// billsFailures makes a failure still cost what was produced, as a truncated answer does.
	billsFailures bool
	asked         []string
}

func (client *fakeClient) Complete(_ context.Context, key secret.Secret, model string, _ Prompt) (Completion, error) {
	client.asked = append(client.asked, model)
	if key.Reveal() != workingKey {
		return Completion{}, fmt.Errorf("%w by the fake provider", ErrInvalidAPIKey)
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

func newTestService(client *fakeClient, store ConfigStore) *Service {
	profiles := []Profile{
		{Provider: ProviderClaude, Client: client, Model: ModelInfo{ID: claudeModel}, Price: Price{InputUSDPerMillion: 1, OutputUSDPerMillion: 5}},
		{Provider: ProviderOpenAI, Client: client, Model: ModelInfo{ID: openAIModel}},
	}
	service := NewService(store, profiles, time.Now, slog.New(slog.DiscardHandler))
	service.pause = func(context.Context, time.Duration) error { return nil }
	return service
}

func savedService(t *testing.T, client *fakeClient) *Service {
	t.Helper()
	service := newTestService(client, &memoryStore{})
	if err := service.Save(t.Context(), Draft{Provider: ProviderClaude, APIKey: workingKey}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return service
}

func TestAFailedAnswerIsNeverAskedOfAnotherModel(t *testing.T) {
	client := &fakeClient{}
	service := savedService(t, client)
	client.failing, client.asked = map[string]error{claudeModel: errors.New("refused")}, nil
	if _, err := service.Complete(t.Context(), Prompt{}); err == nil || strings.Join(client.asked, ",") != claudeModel {
		t.Errorf("Complete = %v after asking %v; want the failure returned after asking only %s", err, client.asked, claudeModel)
	}
}

func TestABusyModelIsAskedAgainThenGivenUp(t *testing.T) {
	client := &fakeClient{}
	service := savedService(t, client)
	client.failing, client.asked = map[string]error{claudeModel: fmt.Errorf("%w: overloaded", ErrTemporary)}, nil
	_, err := service.Complete(t.Context(), Prompt{})
	if !errors.Is(err, ErrTemporary) || len(client.asked) != 1+len(retryPauses) {
		t.Errorf("Complete = %v after asking %v; want ErrTemporary after %d attempts at the same model", err, client.asked, 1+len(retryPauses))
	}
}

func TestAnAnswerIsPricedAtTheProvidersModel(t *testing.T) {
	completion, err := savedService(t, &fakeClient{}).Complete(t.Context(), Prompt{})
	if err != nil || completion.Text != "answer from "+claudeModel || completion.CostUSD != 1+5 {
		t.Errorf("Complete = %+v, %v; want %s's answer priced at $1+$5", completion, err, claudeModel)
	}
}

func TestATruncatedAnswerStillCountsTowardsTheCost(t *testing.T) {
	client := &fakeClient{billsFailures: true}
	service := savedService(t, client)
	client.failing = map[string]error{claudeModel: errors.New("cut off at max tokens")}
	completion, err := service.Complete(t.Context(), Prompt{})
	if err == nil || completion.CostUSD != 1+5 {
		t.Errorf("Complete = %+v, %v; want the failure with its $6 still reported", completion, err)
	}
}

func TestSavingKeepsAWorkingKeyUntilANewOneIsAccepted(t *testing.T) {
	service := savedService(t, &fakeClient{})
	if err := service.Save(t.Context(), Draft{Provider: ProviderClaude, APIKey: "sk-typo"}); !errors.Is(err, ErrInvalidAPIKey) {
		t.Errorf("Save with a rejected key: %v, want ErrInvalidAPIKey", err)
	}
	if err := service.Save(t.Context(), Draft{Provider: ProviderClaude}); err != nil {
		t.Errorf("Save with the key left empty: %v, want the saved key reused", err)
	}
	if err := service.Save(t.Context(), Draft{Provider: ProviderOpenAI}); !errors.Is(err, ErrNoKey) {
		t.Errorf("Save for another provider without a key: %v, want ErrNoKey", err)
	}
	view, err := service.View(t.Context())
	if err != nil || view.Active != ProviderClaude || len(view.Providers) != 1 {
		t.Fatalf("view = %+v, %v; want only Claude saved and in use", view, err)
	}
	if saved := view.Providers[0]; strings.Contains(saved.KeyHint, "working") || !strings.HasSuffix(saved.KeyHint, "1234") {
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

func TestCheckingAsksTheProvidersModel(t *testing.T) {
	client := &fakeClient{}
	service := savedService(t, client)
	client.asked = nil
	check, err := service.Check(t.Context(), ProviderOpenAI, workingKey)
	if err != nil || check.Reply != "answer from "+openAIModel || strings.Join(client.asked, ",") != openAIModel {
		t.Errorf("Check = %+v, %v after asking %v; want %s asked", check, err, client.asked, openAIModel)
	}
}

func TestEachProviderKeepsItsOwnKeyWhenAnotherIsInUse(t *testing.T) {
	service := savedService(t, &fakeClient{})
	if err := service.Save(t.Context(), Draft{Provider: ProviderOpenAI, APIKey: workingKey}); err != nil {
		t.Fatalf("Save OpenAI: %v", err)
	}
	if err := service.Save(t.Context(), Draft{Provider: ProviderClaude}); err != nil {
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

func TestAProviderNoLongerOfferedIsNotUsed(t *testing.T) {
	retired := Provider("gemini")
	store := &memoryStore{found: true, config: Config{
		Active:    retired,
		IsEnabled: true,
		Providers: map[Provider]ProviderSettings{retired: {APIKey: secret.New(workingKey)}, ProviderClaude: {APIKey: secret.New(workingKey)}},
	}}
	service := newTestService(&fakeClient{}, store)
	if _, err := service.Complete(t.Context(), Prompt{}); !errors.Is(err, ErrNotConfigured) {
		t.Errorf("Complete with a retired provider in use: %v, want ErrNotConfigured", err)
	}
	if view, err := service.View(t.Context()); err != nil || view.Active != "" || len(view.Providers) != 1 {
		t.Errorf("view = %+v, %v; want only Claude's key left and nothing in use", view, err)
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
	service := newTestService(&fakeClient{}, &unreadableStore{})
	if view, err := service.View(t.Context()); err != nil || view.Active != "" {
		t.Fatalf("View of unreadable settings = %+v, %v; want empty settings", view, err)
	}
	if err := service.Save(t.Context(), Draft{Provider: ProviderClaude, APIKey: workingKey}); err != nil {
		t.Errorf("Save over unreadable settings: %v", err)
	}
}
