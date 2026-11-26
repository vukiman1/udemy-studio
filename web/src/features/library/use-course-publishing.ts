import { useQueryClient, type UseQueryResult } from '@tanstack/react-query'
import { useEffect } from 'react'
import { usePublishRun, usePublishedCourse } from '../../api/publish-queries'
import { QUERY_KEYS } from '../../api/queries'
import { useStorageSettings } from '../../api/storage-queries'
import type { PublishedCourse } from '../../api/types'
import type { CoursePublish } from './CourseCard'

export interface CoursePublishing {
  readonly isConfigured: boolean
  readonly published: UseQueryResult<PublishedCourse | null>
  readonly of: (courseId: string) => CoursePublish | undefined
}

export function useCoursePublishing(): CoursePublishing {
  const queryClient = useQueryClient()
  const storage = useStorageSettings()
  const isConfigured = storage.data?.configured === true
  const run = usePublishRun()
  const published = usePublishedCourse(isConfigured)
  const latestRun = run.data ?? undefined
  const runState = latestRun?.state

  useEffect(() => {
    if (runState === 'done')
      void queryClient.invalidateQueries({ queryKey: QUERY_KEYS.publishedCourse })
  }, [runState, queryClient])

  return {
    isConfigured,
    published,
    of: (courseId) => {
      if (!isConfigured) return undefined
      return {
        run: latestRun?.courseId === courseId ? latestRun : undefined,
        isPublished: published.data?.courseId === courseId,
      }
    },
  }
}
