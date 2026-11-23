export function coursePath(courseId: string): string {
  return `/courses/${encodeURIComponent(courseId)}`
}

export function lecturePath(courseId: string, lectureId: string): string {
  return `${coursePath(courseId)}/lectures/${encodeURIComponent(lectureId)}`
}
