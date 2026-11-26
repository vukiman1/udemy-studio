import { describe, expect, it } from 'vitest'
import type { PublishPlan } from '../../api/types'
import { initialSelection, selectedBytes } from './chapter-selection'

const COURSE_ID = 'course'

function plan(published: PublishPlan['published']): PublishPlan {
  return {
    sharedBytes: 100,
    limitBytes: 2_000,
    replaces: false,
    published,
    chapters: [
      { dir: '01 - Intro', files: 3, sizeBytes: 1_000, published: true },
      { dir: '02 - Deep', files: 3, sizeBytes: 1_000, published: false },
      { dir: '03 - Extra', files: 3, sizeBytes: 1_000, published: true },
    ],
  }
}

const PUBLISHED_HERE: PublishPlan['published'] = {
  courseId: COURSE_ID,
  dir: 'course',
  title: 'Course',
  sizeBytes: 2_100,
  publishedAt: '2026-11-26T08:00:00Z',
  chapters: 2,
  totalChapters: 3,
}

describe('chapter selection', () => {
  it('starts from the chapters already on R2 when this course is there', () => {
    expect([...initialSelection(plan(PUBLISHED_HERE), COURSE_ID)]).toEqual([
      '01 - Intro',
      '03 - Extra',
    ])
  })

  it('starts from every chapter for a course not on R2', () => {
    expect(initialSelection(plan(undefined), COURSE_ID).size).toBe(3)
  })

  it('counts the shared course files with the chosen chapters', () => {
    expect(selectedBytes(plan(undefined), new Set(['02 - Deep']))).toBe(1_100)
  })
})
