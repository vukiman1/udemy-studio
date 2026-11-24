import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { apiClient } from './client'
import { ApiRequestError, QUERY_KEYS } from './queries'
import type { Note } from './types'

export function useLectureNotes(lectureId: string): UseQueryResult<Note[]> {
  return useQuery({
    queryKey: QUERY_KEYS.notes(lectureId),
    queryFn: async () => {
      const { data, response } = await apiClient.GET('/api/lectures/{lectureId}/notes', {
        params: { path: { lectureId } },
      })
      if (data === undefined)
        throw new ApiRequestError(`load notes of ${lectureId}`, response.status)
      return data
    },
  })
}

export async function addNote(lectureId: string, atSeconds: number, body: string): Promise<Note> {
  const { data, error, response } = await apiClient.POST('/api/lectures/{lectureId}/notes', {
    params: { path: { lectureId } },
    body: { atSeconds, body },
  })
  if (data === undefined)
    throw new ApiRequestError(`add a note to ${lectureId}`, response.status, error.detail)
  return data
}

export async function editNote(noteId: number, body: string): Promise<Note> {
  const { data, error, response } = await apiClient.PATCH('/api/notes/{noteId}', {
    params: { path: { noteId } },
    body: { body },
  })
  if (data === undefined)
    throw new ApiRequestError(`edit note ${String(noteId)}`, response.status, error.detail)
  return data
}

export async function deleteNote(noteId: number): Promise<void> {
  const { error, response } = await apiClient.DELETE('/api/notes/{noteId}', {
    params: { path: { noteId } },
  })
  if (!response.ok)
    throw new ApiRequestError(`delete note ${String(noteId)}`, response.status, error?.detail)
}

export function courseNotesUrl(courseId: string): string {
  return `/api/courses/${encodeURIComponent(courseId)}/notes.md`
}
