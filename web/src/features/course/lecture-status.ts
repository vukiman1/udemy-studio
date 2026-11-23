import type { LectureSummary } from '../../api/types'

export enum LectureStatus {
  Completed = 'completed',
  InProgress = 'in-progress',
  NotStarted = 'not-started',
}

export const LECTURE_STATUS_LABELS: Readonly<Record<LectureStatus, string>> = {
  [LectureStatus.Completed]: 'Đã học',
  [LectureStatus.InProgress]: 'Đang học',
  [LectureStatus.NotStarted]: 'Chưa học',
}

export function lectureStatus(lecture: LectureSummary): LectureStatus {
  if (lecture.progress === undefined) return LectureStatus.NotStarted
  return lecture.progress.completed ? LectureStatus.Completed : LectureStatus.InProgress
}
