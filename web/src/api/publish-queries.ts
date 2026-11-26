import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { apiClient } from './client'
import { ApiRequestError, QUERY_KEYS } from './queries'
import type { PublishPlan, PublishRun, PublishedCourse } from './types'

const ACTIVE_POLL_MS = 2000
const NO_CONTENT = 204

export function isPublishing(run: PublishRun | undefined): boolean {
  return run?.state === 'running'
}

export function usePublishRun(): UseQueryResult<PublishRun | null> {
  return useQuery({
    queryKey: QUERY_KEYS.publishRun,
    refetchInterval: (query) =>
      isPublishing(query.state.data ?? undefined) ? ACTIVE_POLL_MS : false,
    queryFn: async () => {
      const { data, response } = await apiClient.GET('/api/publish')
      if (response.status === NO_CONTENT) return null
      if (data === undefined) throw new ApiRequestError('load the publish run', response.status)
      return data
    },
  })
}

export function usePublishedCourse(isEnabled: boolean): UseQueryResult<PublishedCourse | null> {
  return useQuery({
    queryKey: QUERY_KEYS.publishedCourse,
    enabled: isEnabled,
    queryFn: async () => {
      const { data, error, response } = await apiClient.GET('/api/publish/published')
      if (response.status === NO_CONTENT) return null
      if (data === undefined)
        throw new ApiRequestError('load the published course', response.status, error?.detail)
      return data
    },
  })
}

export function usePublishPlan(courseId: string): UseQueryResult<PublishPlan> {
  return useQuery({
    queryKey: QUERY_KEYS.publishPlan(courseId),
    queryFn: async () => {
      const { data, error, response } = await apiClient.GET(
        '/api/courses/{courseId}/publish/plan',
        {
          params: { path: { courseId } },
        },
      )
      if (data === undefined)
        throw new ApiRequestError(`plan publishing ${courseId}`, response.status, error.detail)
      return data
    },
  })
}

export async function startPublish(
  courseId: string,
  chapters: readonly string[],
): Promise<PublishRun> {
  const { data, error, response } = await apiClient.POST('/api/courses/{courseId}/publish', {
    params: { path: { courseId } },
    body: { chapters: [...chapters] },
  })
  if (data === undefined)
    throw new ApiRequestError(`publish ${courseId}`, response.status, error.detail)
  return data
}

export async function cancelPublish(): Promise<void> {
  const { error, response } = await apiClient.DELETE('/api/publish')
  if (!response.ok) throw new ApiRequestError('stop publishing', response.status, error?.detail)
}
