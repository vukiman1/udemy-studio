import { describe, expect, it } from 'vitest'
import type { LectureSummary } from '../../api/types'
import { LectureStatus, lectureStatus } from './lecture-status'

const BASE: LectureSummary = {
  id: 'l1',
  number: 1,
  title: 'Welcome',
  kind: 'video',
  hasVideo: true,
  hasArticle: false,
  subtitleLanguages: [],
  attachments: [],
}

function withProgress(completed: boolean): LectureSummary {
  return {
    ...BASE,
    progress: {
      positionSeconds: 10,
      durationSeconds: 100,
      completed,
      updatedAt: '2026-11-23T08:00:00Z',
    },
  }
}

describe('lectureStatus', () => {
  it('is not started without any progress', () => {
    expect(lectureStatus(BASE)).toBe(LectureStatus.NotStarted)
  })
  it('is in progress once watched but not finished', () => {
    expect(lectureStatus(withProgress(false))).toBe(LectureStatus.InProgress)
  })
  it('is completed once finished', () => {
    expect(lectureStatus(withProgress(true))).toBe(LectureStatus.Completed)
  })
})
