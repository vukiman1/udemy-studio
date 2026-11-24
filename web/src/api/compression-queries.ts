import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { apiClient } from './client'
import { ApiRequestError, QUERY_KEYS } from './queries'
import type { CompressionPlan, CompressionRun } from './types'

const ACTIVE_POLL_MS = 1500
const NOT_COMPRESSED_STATUS = 204
export const COMPRESSION_UNAVAILABLE_STATUS = 503

export function isCompressing(run: CompressionRun | null | undefined): run is CompressionRun {
  return run?.state === 'queued' || run?.state === 'running'
}

export function useCompressionRun(courseId: string): UseQueryResult<CompressionRun | null> {
  return useQuery({
    queryKey: QUERY_KEYS.compression(courseId),
    refetchInterval: (query) => (isCompressing(query.state.data) ? ACTIVE_POLL_MS : false),
    queryFn: async () => {
      const { data, error, response } = await apiClient.GET('/api/courses/{courseId}/compression', {
        params: { path: { courseId } },
      })
      if (response.status === NOT_COMPRESSED_STATUS) return null
      if (data === undefined)
        throw new ApiRequestError(`load compression of ${courseId}`, response.status, error?.detail)
      return data
    },
  })
}

export function useCompressionPlan(courseId: string): UseQueryResult<CompressionPlan> {
  return useQuery({
    queryKey: QUERY_KEYS.compressionPlan(courseId),
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

export async function startCompression(courseId: string): Promise<CompressionRun> {
  const { data, error, response } = await apiClient.POST('/api/courses/{courseId}/compression', {
    params: { path: { courseId } },
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
