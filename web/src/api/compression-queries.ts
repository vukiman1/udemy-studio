import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { apiClient } from './client'
import { ApiRequestError, QUERY_KEYS } from './queries'
import type { CompressionPlan, CompressionRun } from './types'

export function isCompressing(run: CompressionRun | undefined): boolean {
  return run?.state === 'queued' || run?.state === 'running'
}

export function useCompressionPlan(courseId: string): UseQueryResult<CompressionPlan> {
  return useQuery({
    queryKey: QUERY_KEYS.compressionPlan(courseId),
    gcTime: 0,
    queryFn: async () => {
      const { data, error, response } = await apiClient.GET(
        '/api/courses/{courseId}/compression/plan',
        {
          params: { path: { courseId } },
        },
      )
      if (data === undefined)
        throw new ApiRequestError(`plan compression of ${courseId}`, response.status, error.detail)
      return data
    },
  })
}

export async function startCompression(
  courseId: string,
  chapters: readonly string[],
): Promise<CompressionRun> {
  const { data, error, response } = await apiClient.POST('/api/courses/{courseId}/compression', {
    params: { path: { courseId } },
    body: { chapters: [...chapters] },
  })
  if (data === undefined)
    throw new ApiRequestError(`start compression of ${courseId}`, response.status, error.detail)
  return data
}

export async function cancelCompression(courseId: string): Promise<void> {
  const { error, response } = await apiClient.DELETE('/api/courses/{courseId}/compression', {
    params: { path: { courseId } },
  })
  if (!response.ok)
    throw new ApiRequestError(`cancel compression of ${courseId}`, response.status, error?.detail)
}
