// Package openai answers prompts with OpenAI's Responses API over plain HTTP.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	apiURL = "https://api.openai.com/v1"
	// The cap counts reasoning tokens too, so it leaves room beyond the reply itself.
	maxReplyTokens   = 16000
	requestTimeout   = 10 * time.Minute
	completedStatus  = "completed"
	messageItem      = "message"
	outputTextPart   = "output_text"
	refusalPart      = "refusal"
	textModelPrefix  = "gpt-"
	maxErrorBodySize = 4096
	// Waits longer than this mean a spent quota, so the next model is tried instead.
	maxRetryAfterSeconds = 30
)

var (
	// errForbidden is a key that may not use one model: a rejected key when checking it, but another model may still answer.
	errForbidden  = errors.New("the key may not use this")
	ErrRefused    = errors.New("the model declined to answer")
	ErrIncomplete = errors.New("the model's reply did not complete")
	// Model ids under gpt- that are not text models answering a prompt.
	nonTextModelParts = []string{"audio", "realtime", "image", "tts", "transcribe", "search", "embedding"}
)

type Client struct {
	http    *http.Client
	baseURL string
}

// Prices are per million tokens on the standard tier, read from OpenAI's pricing page on 2026-11-24.
func Profile() llm.Profile {
	return llm.Profile{
		Provider: llm.ProviderOpenAI,
		Label:    "ChatGPT (OpenAI)",
		Known: []llm.ModelInfo{
			{ID: "gpt-6.2-astra", Label: "GPT-6.2 Astra"},
			{ID: "gpt-6.2-sol", Label: "GPT-6.2 Sol"},
			{ID: "gpt-6.1-terra", Label: "GPT-6.1 Terra"},
			{ID: "gpt-6-luna", Label: "GPT-6 Luna"},
		},
		// Measured on a 303-cue lecture: Terra matched Sol's quality at half the cost.
		Recommended: []string{"gpt-6.1-terra", "gpt-6.2-sol"},
		Prices: map[string]llm.Price{
			"gpt-6.2-astra": {InputUSDPerMillion: 8, OutputUSDPerMillion: 40},
			"gpt-6.2-sol":   {InputUSDPerMillion: 2, OutputUSDPerMillion: 10},
			"gpt-6.1-terra": {InputUSDPerMillion: 1, OutputUSDPerMillion: 5},
			"gpt-6-luna":    {InputUSDPerMillion: 0.10, OutputUSDPerMillion: 0.50},
		},
		Client: Client{http: &http.Client{Timeout: requestTimeout}, baseURL: apiURL},
	}
}

type modelList struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (client Client) CheckKey(ctx context.Context, key secret.Secret) error {
	_, err := client.ListModels(ctx, key)
	if errors.Is(err, errForbidden) {
		return fmt.Errorf("%w: %w", llm.ErrInvalidAPIKey, err)
	}
	return err
}

func (client Client) ListModels(ctx context.Context, key secret.Secret) ([]llm.ModelInfo, error) {
	var listed modelList
	if err := client.call(ctx, key, http.MethodGet, "/models", nil, &listed); err != nil {
		return nil, fmt.Errorf("list OpenAI models: %w", err)
	}
	var models []llm.ModelInfo
	for _, model := range listed.Data {
		if isTextModel(model.ID) {
			models = append(models, llm.ModelInfo{ID: model.ID, Label: model.ID})
		}
	}
	slices.SortFunc(models, func(left, right llm.ModelInfo) int { return strings.Compare(left.ID, right.ID) })
	return models, nil
}

type responseRequest struct {
	Model           string `json:"model"`
	Instructions    string `json:"instructions"`
	Input           string `json:"input"`
	MaxOutputTokens int    `json:"max_output_tokens"`
	Store           bool   `json:"store"`
}

// The usage field names follow the Responses API; a reply without them is still used, just not priced.
type response struct {
	Status string `json:"status"`
	Output []struct {
		Type    string `json:"type"`
		Content []struct {
			Type    string `json:"type"`
			Text    string `json:"text"`
			Refusal string `json:"refusal"`
		} `json:"content"`
	} `json:"output"`
	Usage struct {
		InputTokens  int64 `json:"input_tokens"`
		OutputTokens int64 `json:"output_tokens"`
	} `json:"usage"`
}

func (client Client) Complete(ctx context.Context, key secret.Secret, model string, prompt llm.Prompt) (llm.Completion, error) {
	request := responseRequest{Model: model, Instructions: prompt.System, Input: prompt.User, MaxOutputTokens: maxReplyTokens}
	var reply response
	if err := client.call(ctx, key, http.MethodPost, "/responses", request, &reply); err != nil {
		return llm.Completion{}, fmt.Errorf("call %s: %w", model, err)
	}
	usage := llm.Usage{InputTokens: reply.Usage.InputTokens, OutputTokens: reply.Usage.OutputTokens}
	var text strings.Builder
	for _, item := range reply.Output {
		if item.Type != messageItem {
			continue
		}
		for _, part := range item.Content {
			switch part.Type {
			case outputTextPart:
				text.WriteString(part.Text)
			case refusalPart:
				return llm.Completion{Usage: usage}, fmt.Errorf("%w: %s", ErrRefused, part.Refusal)
			}
		}
	}
	if reply.Status != "" && reply.Status != completedStatus {
		return llm.Completion{Usage: usage}, fmt.Errorf("%w (status %s)", ErrIncomplete, reply.Status)
	}
	return llm.Completion{Text: text.String(), Usage: usage}, nil
}

type errorBody struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (client Client) call(ctx context.Context, key secret.Secret, method, path string, body, into any) error {
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		payload = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, client.baseURL+path, payload)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+key.Reveal())
	request.Header.Set("Content-Type", "application/json")
	response, err := client.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return statusError(response)
	}
	if err := json.NewDecoder(response.Body).Decode(into); err != nil {
		return fmt.Errorf("read OpenAI reply: %w", err)
	}
	return nil
}

func statusError(response *http.Response) error {
	raw, _ := io.ReadAll(io.LimitReader(response.Body, maxErrorBodySize))
	var parsed errorBody
	message := strings.TrimSpace(string(raw))
	if json.Unmarshal(raw, &parsed) == nil && parsed.Error.Message != "" {
		message = parsed.Error.Message
	}
	switch response.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("%w by OpenAI (HTTP %d): %s", llm.ErrInvalidAPIKey, response.StatusCode, message)
	case http.StatusForbidden:
		return fmt.Errorf("%w (HTTP %d): %s", errForbidden, response.StatusCode, message)
	}
	if isTemporary(response) {
		return fmt.Errorf("%w: OpenAI answered HTTP %d: %s", llm.ErrTemporary, response.StatusCode, message)
	}
	return fmt.Errorf("OpenAI answered HTTP %d: %s", response.StatusCode, message)
}

// isTemporary is true for an overloaded server, and for a rate limit only when it asks for a short wait;
// a spent daily quota is not worth waiting for.
func isTemporary(response *http.Response) bool {
	switch response.StatusCode {
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	case http.StatusTooManyRequests:
		seconds, err := strconv.Atoi(response.Header.Get("Retry-After"))
		return err == nil && seconds <= maxRetryAfterSeconds
	}
	return false
}

func isTextModel(id string) bool {
	if !strings.HasPrefix(id, textModelPrefix) {
		return false
	}
	return !slices.ContainsFunc(nonTextModelParts, func(part string) bool { return strings.Contains(id, part) })
}
