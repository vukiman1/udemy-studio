import { useQueryClient, type UseQueryResult } from '@tanstack/react-query'
import { useEffect } from 'react'
import { usePublishRun, usePublishedCourse } from '../../api/publish-queries'
import { QUERY_KEYS } from '../../api/queries'
import { useStorageSettings, useStorageUsage } from '../../api/storage-queries'
import type { PublishedCourse, StorageUsage } from '../../api/types'
import type { CoursePublish } from './CourseCard'

export interface CoursePublishing {
  readonly isConfigured: boolean
  readonly published: UseQueryResult<PublishedCourse | null>
  readonly usage: UseQueryResult<StorageUsage | null>
  readonly of: (courseId: string) => CoursePublish | undefined
}

export function useCoursePublishing(isLibraryCopy: boolean): CoursePublishing {
  const queryClient = useQueryClient()
  const storage = useStorageSettings()
  const isConfigured = isLibraryCopy && storage.data?.configured === true
  const run = usePublishRun(isLibraryCopy)
  const published = usePublishedCourse(isConfigured)
  const usage = useStorageUsage(isConfigured)
  const latestRun = run.data ?? undefined
  const runState = latestRun?.state

  useEffect(() => {
    if (runState === undefined || runState === 'running') return
    void queryClient.invalidateQueries({ queryKey: QUERY_KEYS.publishedCourse })
    void queryClient.invalidateQueries({ queryKey: QUERY_KEYS.storageUsage })
  }, [runState, queryClient])

  return {
    isConfigured,
    published,
    usage,
    of: (courseId) => {
      if (!isConfigured) return undefined
      return {
        run: latestRun?.courseId === courseId ? latestRun : undefined,
        isPublished: published.data?.courseId === courseId,
      }
    },
  }
}
