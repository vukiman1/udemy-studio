// Package gemini answers prompts with the Gemini API's generateContent over plain HTTP.
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	apiURL           = "https://generativelanguage.googleapis.com/v1beta"
	apiKeyHeader     = "x-goog-api-key" //nolint:gosec // reason: the name of the header, not a credential
	maxReplyTokens   = 16000
	requestTimeout   = 10 * time.Minute
	modelNamePrefix  = "models/"
	generateMethod   = "generateContent"
	listPageSize     = "1000"
	finishedReason   = "STOP"
	maxErrorBodySize = 4096
	// Waits longer than this mean a spent quota, so the next model is tried instead.
	maxRetryAfterSeconds = 30
)

var (
	// errForbidden is a key that may not use one model: a rejected key when checking it, but another model may still answer.
	errForbidden  = errors.New("the key may not use this")
	ErrBlocked    = errors.New("gemini declined to answer")
	ErrUnfinished = errors.New("gemini's reply did not finish")
	// Google reports a bad key as HTTP 400 with one of these markers rather than a 401.
	invalidKeyMarkers = []string{"API_KEY_INVALID", "API key not valid"}
)

type Client struct {
	http    *http.Client
	baseURL string
}

// Prices are per million tokens on the paid tier, read from the Gemini pricing page on 2026-11-24;
// Gemini 4 Carbon rises to $4/$20 on 2026-12-01.
func Profile() llm.Profile {
	return llm.Profile{
		Provider: llm.ProviderGemini,
		Label:    "Gemini (Google)",
		Known: []llm.ModelInfo{
			{ID: "gemini-4-carbon", Label: "Gemini 4 Carbon"},
			{ID: "gemini-3.8-flash", Label: "Gemini 3.8 Flash"},
			{ID: "gemini-3.5-flash-lite", Label: "Gemini 3.5 Flash-Lite"},
		},
		Recommended: []string{"gemini-4-carbon", "gemini-3.8-flash"},
		Prices: map[string]llm.Price{
			"gemini-4-carbon":       {InputUSDPerMillion: 2, OutputUSDPerMillion: 10},
			"gemini-3.8-flash":      {InputUSDPerMillion: 0.75, OutputUSDPerMillion: 3.75},
			"gemini-3.5-flash-lite": {InputUSDPerMillion: 0.30, OutputUSDPerMillion: 2.50},
		},
		Client: Client{http: &http.Client{Timeout: requestTimeout}, baseURL: apiURL},
	}
}

type modelPage struct {
	Models []struct {
		Name                       string   `json:"name"`
		DisplayName                string   `json:"displayName"`
		SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
	} `json:"models"`
	NextPageToken string `json:"nextPageToken"`
}

func (client Client) CheckKey(ctx context.Context, key secret.Secret) error {
	_, err := client.ListModels(ctx, key)
	if errors.Is(err, errForbidden) {
		return fmt.Errorf("%w: %w", llm.ErrInvalidAPIKey, err)
	}
	return err
}

func (client Client) ListModels(ctx context.Context, key secret.Secret) ([]llm.ModelInfo, error) {
	var models []llm.ModelInfo
	pageToken := ""
	for {
		query := url.Values{"pageSize": {listPageSize}}
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		var page modelPage
		if err := client.call(ctx, key, http.MethodGet, "/models?"+query.Encode(), nil, &page); err != nil {
			return nil, fmt.Errorf("list Gemini models: %w", err)
		}
		for _, model := range page.Models {
			if slices.Contains(model.SupportedGenerationMethods, generateMethod) {
				models = append(models, llm.ModelInfo{ID: strings.TrimPrefix(model.Name, modelNamePrefix), Label: model.DisplayName})
			}
		}
		if page.NextPageToken == "" {
			return models, nil
		}
		pageToken = page.NextPageToken
	}
}

type part struct {
	Text    string `json:"text,omitempty"`
	Thought bool   `json:"thought,omitempty"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type generateRequest struct {
	SystemInstruction content          `json:"systemInstruction"`
	Contents          []content        `json:"contents"`
	GenerationConfig  generationConfig `json:"generationConfig"`
}

type generationConfig struct {
	MaxOutputTokens int `json:"maxOutputTokens"`
}

type generateResponse struct {
	Candidates []struct {
		Content      content `json:"content"`
		FinishReason string  `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
	UsageMetadata struct {
		PromptTokenCount     int64 `json:"promptTokenCount"`
		CandidatesTokenCount int64 `json:"candidatesTokenCount"`
		ThoughtsTokenCount   int64 `json:"thoughtsTokenCount"`
	} `json:"usageMetadata"`
}

func (client Client) Complete(ctx context.Context, key secret.Secret, model string, prompt llm.Prompt) (llm.Completion, error) {
	request := generateRequest{
		SystemInstruction: content{Parts: []part{{Text: prompt.System}}},
		Contents:          []content{{Role: "user", Parts: []part{{Text: prompt.User}}}},
		GenerationConfig:  generationConfig{MaxOutputTokens: maxReplyTokens},
	}
	var reply generateResponse
	if err := client.call(ctx, key, http.MethodPost, "/models/"+url.PathEscape(model)+":"+generateMethod, request, &reply); err != nil {
		return llm.Completion{}, fmt.Errorf("call %s: %w", model, err)
	}
	metadata := reply.UsageMetadata
	// Thinking tokens are billed as output.
	usage := llm.Usage{InputTokens: metadata.PromptTokenCount, OutputTokens: metadata.CandidatesTokenCount + metadata.ThoughtsTokenCount}
	if reply.PromptFeedback.BlockReason != "" {
		return llm.Completion{Usage: usage}, fmt.Errorf("%w (%s)", ErrBlocked, reply.PromptFeedback.BlockReason)
	}
	if len(reply.Candidates) == 0 {
		return llm.Completion{Usage: usage}, fmt.Errorf("%w: no candidate", ErrUnfinished)
	}
	candidate := reply.Candidates[0]
	if candidate.FinishReason != finishedReason {
		return llm.Completion{Usage: usage}, fmt.Errorf("%w (%s)", ErrUnfinished, candidate.FinishReason)
	}
	var text strings.Builder
	for _, piece := range candidate.Content.Parts {
		if !piece.Thought {
			text.WriteString(piece.Text)
		}
	}
	return llm.Completion{Text: text.String(), Usage: usage}, nil
}

type errorBody struct {
	Error struct {
		Message string `json:"message"`
		Status  string `json:"status"`
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
	request.Header.Set(apiKeyHeader, key.Reveal())
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
		return fmt.Errorf("read Gemini reply: %w", err)
	}
	return nil
}

func statusError(response *http.Response) error {
	raw, _ := io.ReadAll(io.LimitReader(response.Body, maxErrorBodySize))
	message := strings.TrimSpace(string(raw))
	var parsed errorBody
	if json.Unmarshal(raw, &parsed) == nil && parsed.Error.Message != "" {
		message = parsed.Error.Message
	}
	isRejectedKey := response.StatusCode == http.StatusUnauthorized ||
		slices.ContainsFunc(invalidKeyMarkers, func(marker string) bool { return strings.Contains(string(raw), marker) })
	if isRejectedKey {
		return fmt.Errorf("%w by Gemini (HTTP %d): %s", llm.ErrInvalidAPIKey, response.StatusCode, message)
	}
	if response.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%w (HTTP %d): %s", errForbidden, response.StatusCode, message)
	}
	if isTemporary(response) {
		return fmt.Errorf("%w: gemini answered HTTP %d: %s", llm.ErrTemporary, response.StatusCode, message)
	}
	return fmt.Errorf("gemini answered HTTP %d: %s", response.StatusCode, message)
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
