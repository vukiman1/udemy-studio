import type { CourseDetail, LectureSummary } from '../../api/types'

export interface LectureNeighbours {
  readonly previous: LectureSummary | undefined
  readonly next: LectureSummary | undefined
}

export function courseLectures(course: CourseDetail): LectureSummary[] {
  return course.chapters.flatMap((chapter) => chapter.lectures)
}

export function findLecture(course: CourseDetail, lectureId: string): LectureSummary | undefined {
  return courseLectures(course).find((lecture) => lecture.id === lectureId)
}

export function neighboursOf(course: CourseDetail, lectureId: string): LectureNeighbours {
  const lectures = courseLectures(course)
  const position = lectures.findIndex((lecture) => lecture.id === lectureId)
  if (position < 0) return { previous: undefined, next: undefined }
  return { previous: lectures[position - 1], next: lectures[position + 1] }
}

export function startingLectureId(course: CourseDetail): string | undefined {
  if (course.resume !== undefined) return course.resume.lectureId
  return courseLectures(course).find((lecture) => lecture.kind !== 'unavailable')?.id
}
