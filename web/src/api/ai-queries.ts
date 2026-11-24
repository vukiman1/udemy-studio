import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { apiClient } from './client'
import { ApiRequestError, QUERY_KEYS } from './queries'
import type { AICheck, AIModel, AIProvider, AIProviderId, AISettings } from './types'

export function useAIProviders(): UseQueryResult<AIProvider[]> {
  return useQuery({
    queryKey: QUERY_KEYS.aiProviders,
    staleTime: Infinity,
    queryFn: async () => {
      const { data, response } = await apiClient.GET('/api/ai/providers')
      if (data === undefined) throw new ApiRequestError('list AI providers', response.status)
      return data
    },
  })
}

export function useAISettings(): UseQueryResult<AISettings> {
  return useQuery({
    queryKey: QUERY_KEYS.aiSettings,
    queryFn: async () => {
      const { data, response } = await apiClient.GET('/api/ai/settings')
      if (data === undefined) throw new ApiRequestError('load AI settings', response.status)
      return data
    },
  })
}

export interface AIDraft {
  readonly provider: AIProviderId
  // Empty keeps the saved key of the same provider.
  readonly apiKey: string
  readonly models: readonly string[]
}

function keyField(apiKey: string): { apiKey?: string } {
  return apiKey.trim() === '' ? {} : { apiKey }
}

export async function saveAISettings(draft: AIDraft): Promise<AISettings> {
  const { data, error, response } = await apiClient.PUT('/api/ai/settings', {
    body: { provider: draft.provider, models: [...draft.models], ...keyField(draft.apiKey) },
  })
  if (data === undefined)
    throw new ApiRequestError('save AI settings', response.status, error.detail)
  return data
}

export async function setAIEnabled(enabled: boolean): Promise<AISettings> {
  const { data, error, response } = await apiClient.PUT('/api/ai/settings/enabled', {
    body: { enabled },
  })
  if (data === undefined)
    throw new ApiRequestError('change whether AI is used', response.status, error.detail)
  return data
}

export async function forgetAIProvider(provider: AIProviderId): Promise<void> {
  const { response } = await apiClient.DELETE('/api/ai/settings/providers/{provider}', {
    params: { path: { provider } },
  })
  if (!response.ok) throw new ApiRequestError(`forget the ${provider} key`, response.status)
}

// With a model the check also calls it with a tiny prompt; without one it only checks the key.
export async function checkAI(
  provider: AIProviderId,
  apiKey: string,
  model: string | undefined,
): Promise<AICheck> {
  const { data, error, response } = await apiClient.POST('/api/ai/check', {
    body: { provider, ...keyField(apiKey), ...(model === undefined ? {} : { model }) },
  })
  if (data === undefined)
    throw new ApiRequestError('check the API key', response.status, error.detail)
  return data
}

export async function listAIModels(provider: AIProviderId, apiKey: string): Promise<AIModel[]> {
  const { data, error, response } = await apiClient.POST('/api/ai/models', {
    body: { provider, ...keyField(apiKey) },
  })
  if (data === undefined) throw new ApiRequestError('list models', response.status, error.detail)
  return data
}
