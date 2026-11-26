export function coursePath(courseId: string): string {
  return `/courses/${encodeURIComponent(courseId)}`
}

export function lecturePath(courseId: string, lectureId: string): string {
  return `${coursePath(courseId)}/lectures/${encodeURIComponent(lectureId)}`
}

export const LIBRARY_PATH = '/'
export const DOWNLOADS_PATH = '/downloads'
export const SEARCH_PATH = '/search'
export const HISTORY_PATH = '/history'
export const SEARCH_QUERY_PARAM = 'q'
// The lecture page starts the video here instead of where it was left.
export const START_SECONDS_PARAM = 't'

export function lectureAtPath(courseId: string, lectureId: string, startSeconds: number): string {
  const query = new URLSearchParams({ [START_SECONDS_PARAM]: String(startSeconds) })
  return `${lecturePath(courseId, lectureId)}?${query.toString()}`
}

export const AI_SETTINGS_PATH = '/settings/ai'
export const STORAGE_SETTINGS_PATH = '/settings/storage'
