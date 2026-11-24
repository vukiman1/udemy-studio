package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func (server apiServer) ListAIProviders(_ context.Context, _ gen.ListAIProvidersRequestObject) (gen.ListAIProvidersResponseObject, error) {
	profiles := server.deps.AI.Profiles()
	response := make(gen.ListAIProviders200JSONResponse, len(profiles))
	for position, profile := range profiles {
		response[position] = gen.AIProvider{
			Id:          gen.AIProviderId(profile.Provider),
			Label:       profile.Label,
			Models:      toAIModels(profile.Known),
			Recommended: profile.Recommended,
		}
	}
	return response, nil
}

func (server apiServer) GetAISettings(ctx context.Context, _ gen.GetAISettingsRequestObject) (gen.GetAISettingsResponseObject, error) {
	settings, err := server.aiSettings(ctx)
	if err != nil {
		return nil, err
	}
	return gen.GetAISettings200JSONResponse(settings), nil
}

func (server apiServer) SaveAISettings(ctx context.Context, request gen.SaveAISettingsRequestObject) (gen.SaveAISettingsResponseObject, error) {
	draft := llm.Draft{Provider: llm.Provider(request.Body.Provider), APIKey: valueOf(request.Body.ApiKey), Models: request.Body.Models}
	err := server.deps.AI.Save(ctx, draft)
	switch {
	case isBadAIRequest(err):
		return gen.SaveAISettings400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	case errors.Is(err, llm.ErrProvider):
		return gen.SaveAISettings502ApplicationProblemPlusJSONResponse{ProviderUnreachableApplicationProblemPlusJSONResponse: providerUnreachable(err)}, nil
	case err != nil:
		return nil, err
	}
	settings, err := server.aiSettings(ctx)
	if err != nil {
		return nil, err
	}
	return gen.SaveAISettings200JSONResponse(settings), nil
}

func (server apiServer) ForgetAIProvider(ctx context.Context, request gen.ForgetAIProviderRequestObject) (gen.ForgetAIProviderResponseObject, error) {
	if err := server.deps.AI.Forget(ctx, llm.Provider(request.Provider)); err != nil {
		return nil, err
	}
	return gen.ForgetAIProvider204Response{}, nil
}

func (server apiServer) SetAIEnabled(ctx context.Context, request gen.SetAIEnabledRequestObject) (gen.SetAIEnabledResponseObject, error) {
	err := server.deps.AI.SetEnabled(ctx, request.Body.Enabled)
	if errors.Is(err, llm.ErrNotConfigured) {
		return gen.SetAIEnabled409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict(err)}, nil
	}
	if err != nil {
		return nil, err
	}
	settings, err := server.aiSettings(ctx)
	if err != nil {
		return nil, err
	}
	return gen.SetAIEnabled200JSONResponse(settings), nil
}

func (server apiServer) CheckAI(ctx context.Context, request gen.CheckAIRequestObject) (gen.CheckAIResponseObject, error) {
	check, err := server.deps.AI.Check(ctx, llm.Provider(request.Body.Provider), valueOf(request.Body.ApiKey), valueOf(request.Body.Model))
	switch {
	case err == nil:
		response := gen.CheckAI200JSONResponse{}
		if check.Model != "" {
			latencyMs := check.Latency.Milliseconds()
			response.Model, response.LatencyMs, response.Reply = &check.Model, &latencyMs, &check.Reply
		}
		return response, nil
	case isBadAIRequest(err):
		return gen.CheckAI400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	case errors.Is(err, llm.ErrProvider):
		return gen.CheckAI502ApplicationProblemPlusJSONResponse{ProviderUnreachableApplicationProblemPlusJSONResponse: providerUnreachable(err)}, nil
	default:
		return nil, err
	}
}

func (server apiServer) ListAIModels(ctx context.Context, request gen.ListAIModelsRequestObject) (gen.ListAIModelsResponseObject, error) {
	models, err := server.deps.AI.ListModels(ctx, llm.Provider(request.Body.Provider), valueOf(request.Body.ApiKey))
	switch {
	case err == nil:
		return gen.ListAIModels200JSONResponse(toAIModels(models)), nil
	case isBadAIRequest(err):
		return gen.ListAIModels400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	case errors.Is(err, llm.ErrProvider):
		return gen.ListAIModels502ApplicationProblemPlusJSONResponse{ProviderUnreachableApplicationProblemPlusJSONResponse: providerUnreachable(err)}, nil
	default:
		return nil, err
	}
}

func (server apiServer) aiSettings(ctx context.Context) (gen.AISettings, error) {
	view, err := server.deps.AI.View(ctx)
	if err != nil {
		return gen.AISettings{}, err
	}
	settings := gen.AISettings{Configured: view.Active != "", Enabled: view.IsEnabled, Providers: make([]gen.AISavedProvider, len(view.Providers))}
	for position, saved := range view.Providers {
		settings.Providers[position] = gen.AISavedProvider{Provider: gen.AIProviderId(saved.Provider), KeyHint: saved.KeyHint, Models: saved.Models}
	}
	if settings.Configured {
		active := gen.AIProviderId(view.Active)
		settings.Provider = &active
	}
	return settings, nil
}

// isBadAIRequest separates what the learner can fix in the form from a provider that could not be reached.
func isBadAIRequest(err error) bool {
	return errors.Is(err, llm.ErrInvalidAPIKey) || errors.Is(err, llm.ErrNoKey) || errors.Is(err, llm.ErrNoModels) || errors.Is(err, llm.ErrUnknownProvider)
}

func providerUnreachable(err error) gen.ProviderUnreachableApplicationProblemPlusJSONResponse {
	return gen.ProviderUnreachableApplicationProblemPlusJSONResponse(problem(http.StatusBadGateway, err.Error()))
}

func toAIModels(models []llm.ModelInfo) []gen.AIModel {
	converted := make([]gen.AIModel, len(models))
	for position, model := range models {
		converted[position] = gen.AIModel{Id: model.ID, Label: model.Label}
	}
	return converted
}

func valueOf(optional *string) string {
	if optional == nil {
		return ""
	}
	return *optional
}
