import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { apiClient } from './client'
import type { components } from './schema'
import type {
  ContinueTarget,
  CourseDetail,
  CourseSummary,
  Cue,
  LectureProgress,
  SearchHit,
} from './types'

type ProgressReport = components['schemas']['ProgressReport']

export class ApiRequestError extends Error {
  readonly status: number

  constructor(operation: string, status: number, detail?: string) {
    super(detail ?? `${operation} failed with HTTP ${String(status)}`)
    this.name = 'ApiRequestError'
    this.status = status
  }
}

export const QUERY_KEYS = {
  courses: ['courses'],
  course: (courseId: string) => ['course', courseId],
  continueTarget: ['continue'],
  cues: (lectureId: string, languages: readonly string[]) => [
    'cues',
    lectureId,
    languages.join(','),
  ],
  chromeProfiles: ['chrome-profiles'],
  loginStatus: ['login-status'],
  jobs: ['jobs'],
  search: (query: string) => ['search', query],
  notes: (lectureId: string) => ['notes', lectureId],
  compression: (courseId: string) => ['compression', courseId],
  compressionPlan: (courseId: string) => ['compression-plan', courseId],
} as const

export function useCourses(): UseQueryResult<CourseSummary[]> {
  return useQuery({
    queryKey: QUERY_KEYS.courses,
    queryFn: async () => {
      const { data, response } = await apiClient.GET('/api/courses')
      if (data === undefined) throw new ApiRequestError('list courses', response.status)
      return data
    },
  })
}

export function useCourse(courseId: string): UseQueryResult<CourseDetail> {
  return useQuery({
    queryKey: QUERY_KEYS.course(courseId),
    queryFn: async () => {
      const { data, response } = await apiClient.GET('/api/courses/{courseId}', {
        params: { path: { courseId } },
      })
      if (data === undefined) throw new ApiRequestError(`load course ${courseId}`, response.status)
      return data
    },
  })
}

export function useContinueTarget(): UseQueryResult<ContinueTarget | null> {
  return useQuery({
    queryKey: QUERY_KEYS.continueTarget,
    queryFn: async () => {
      const { data, response } = await apiClient.GET('/api/continue')
      if (response.status === 204) return null
      if (data === undefined) throw new ApiRequestError('load continue target', response.status)
      return data
    },
  })
}

export function useCues(lectureId: string, languages: readonly string[]): UseQueryResult<Cue[]> {
  return useQuery({
    queryKey: QUERY_KEYS.cues(lectureId, languages),
    enabled: languages.length > 0,
    staleTime: Infinity,
    queryFn: async () => {
      const { data, response } = await apiClient.GET('/api/lectures/{lectureId}/cues', {
        params: { path: { lectureId }, query: { langs: [...languages] } },
      })
      if (data === undefined)
        throw new ApiRequestError(`load cues of ${lectureId}`, response.status)
      return data
    },
  })
}

export function useSubtitleSearch(query: string): UseQueryResult<SearchHit[]> {
  return useQuery({
    queryKey: QUERY_KEYS.search(query),
    enabled: query !== '',
    queryFn: async () => {
      const { data, response } = await apiClient.GET('/api/search', {
        params: { query: { q: query } },
      })
      if (data === undefined) throw new ApiRequestError(`search for ${query}`, response.status)
      return data
    },
  })
}

export async function saveProgress(
  lectureId: string,
  report: ProgressReport,
  shouldKeepAlive: boolean,
): Promise<LectureProgress> {
  const { data, response } = await apiClient.PUT('/api/lectures/{lectureId}/progress', {
    params: { path: { lectureId } },
    body: report,
    keepalive: shouldKeepAlive,
  })
  if (data === undefined)
    throw new ApiRequestError(`save progress of ${lectureId}`, response.status)
  return data
}

export function subtitlesVttUrl(lectureId: string, languages: readonly string[]): string {
  const query = new URLSearchParams({ langs: languages.join(',') })
  return `/api/lectures/${encodeURIComponent(lectureId)}/subtitles.vtt?${query.toString()}`
}

export function videoUrl(lectureId: string): string {
  return `/media/${encodeURIComponent(lectureId)}`
}

export function articleUrl(lectureId: string): string {
  return `/articles/${encodeURIComponent(lectureId)}`
}

export function attachmentUrl(attachmentId: string): string {
  return `/attachments/${encodeURIComponent(attachmentId)}`
}
