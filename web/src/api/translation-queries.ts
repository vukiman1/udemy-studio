import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { apiClient } from './client'
import { ApiRequestError, QUERY_KEYS } from './queries'
import type { TranslationPlan, TranslationRun } from './types'

export function isTranslating(run: TranslationRun | undefined): boolean {
  return run?.state === 'queued' || run?.state === 'running'
}

export function useTranslationPlan(courseId: string): UseQueryResult<TranslationPlan> {
  return useQuery({
    queryKey: QUERY_KEYS.translationPlan(courseId),
    queryFn: async () => {
      const { data, error, response } = await apiClient.GET(
        '/api/courses/{courseId}/translation/plan',
        { params: { path: { courseId } } },
      )
      if (data === undefined)
        throw new ApiRequestError(`plan translation of ${courseId}`, response.status, error.detail)
      return data
    },
  })
}

export async function startTranslation(courseId: string): Promise<TranslationRun> {
  const { data, error, response } = await apiClient.POST('/api/courses/{courseId}/translation', {
    params: { path: { courseId } },
  })
  if (data === undefined)
    throw new ApiRequestError(`start translation of ${courseId}`, response.status, error.detail)
  return data
}

export async function cancelTranslation(courseId: string): Promise<void> {
  const { error, response } = await apiClient.DELETE('/api/courses/{courseId}/translation', {
    params: { path: { courseId } },
  })
  if (!response.ok)
    throw new ApiRequestError(`cancel translation of ${courseId}`, response.status, error?.detail)
}
