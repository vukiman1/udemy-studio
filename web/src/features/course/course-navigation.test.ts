import { describe, expect, it } from 'vitest'
import type { CourseDetail, LectureSummary } from '../../api/types'
import { neighboursOf, startingLectureId } from './course-navigation'

function lecture(id: string, kind: LectureSummary['kind'] = 'video'): LectureSummary {
  return {
    id,
    number: 1,
    title: id,
    kind,
    hasVideo: kind === 'video',
    hasArticle: kind === 'article',
    subtitleLanguages: [],
    attachments: [],
  }
}

function course(overrides: Partial<CourseDetail> = {}): CourseDetail {
  return {
    id: 'c',
    title: 'Course',
    playableLectures: 3,
    completedLectures: 0,
    chapters: [
      { index: 1, title: 'One', lectures: [lecture('drm', 'unavailable'), lecture('a')] },
      { index: 2, title: 'Two', lectures: [lecture('b'), lecture('c', 'article')] },
    ],
    ...overrides,
  }
}

describe('neighboursOf', () => {
  it('crosses chapter boundaries', () => {
    const { previous, next } = neighboursOf(course(), 'a')
    expect(previous?.id).toBe('drm')
    expect(next?.id).toBe('b')
  })

  it('has no next lecture at the end', () => {
    expect(neighboursOf(course(), 'c').next).toBeUndefined()
  })

  it('has no neighbours for an unknown lecture', () => {
    expect(neighboursOf(course(), 'zzz')).toEqual({ previous: undefined, next: undefined })
  })
})

describe('startingLectureId', () => {
  it('prefers where the learner left off', () => {
    expect(
      startingLectureId(course({ resume: { courseId: 'c', lectureId: 'b', positionSeconds: 3 } })),
    ).toBe('b')
  })

  it('otherwise opens the first playable lecture', () => {
    expect(startingLectureId(course())).toBe('a')
  })
})
