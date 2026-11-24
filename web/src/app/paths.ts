export function coursePath(courseId: string): string {
  return `/courses/${encodeURIComponent(courseId)}`
}

export function lecturePath(courseId: string, lectureId: string): string {
  return `${coursePath(courseId)}/lectures/${encodeURIComponent(lectureId)}`
}

export const SEARCH_PATH = '/search'
export const SEARCH_QUERY_PARAM = 'q'
// The lecture page starts the video here instead of where it was left.
export const START_SECONDS_PARAM = 't'

export function lectureAtPath(courseId: string, lectureId: string, startSeconds: number): string {
  const query = new URLSearchParams({ [START_SECONDS_PARAM]: String(startSeconds) })
  return `${lecturePath(courseId, lectureId)}?${query.toString()}`
}
