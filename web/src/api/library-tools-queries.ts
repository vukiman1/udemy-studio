import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { apiClient } from './client'
import { isCompressing } from './compression-queries'
import { ApiRequestError, QUERY_KEYS } from './queries'
import { isTranslating } from './translation-queries'
import type { CourseTools } from './types'

const ACTIVE_POLL_MS = 2000

// One query feeds every course card, polled only while something is translating or compressing.
export function useCourseTools(): UseQueryResult<CourseTools[]> {
  return useQuery({
    queryKey: QUERY_KEYS.courseTools,
    refetchInterval: (query) => (query.state.data?.some(isBusy) === true ? ACTIVE_POLL_MS : false),
    queryFn: async () => {
      const { data, response } = await apiClient.GET('/api/library/tools')
      if (data === undefined) throw new ApiRequestError('load course tools', response.status)
      return data
    },
  })
}

function isBusy(tools: CourseTools): boolean {
  return (
    isTranslating(tools.translation.run) ||
    isCompressing(tools.compression?.run) ||
    (tools.compression?.unchecked ?? 0) > 0
  )
}
