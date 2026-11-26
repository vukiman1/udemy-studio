import { describe, expect, it } from 'vitest'
import type { PublishPlan } from '../../api/types'
import { hasChanges, initialSelection, selectedBytes } from './chapter-selection'

const COURSE_ID = 'course'

function plan(published: PublishPlan['published']): PublishPlan {
  return {
    sharedBytes: 100,
    limitBytes: 2_000,
    replaces: false,
    sharedChanged: false,
    published,
    chapters: [
      { dir: '01 - Intro', files: 3, sizeBytes: 1_000, published: true, changed: false },
      { dir: '02 - Deep', files: 3, sizeBytes: 1_000, published: false, changed: true },
      { dir: '03 - Extra', files: 3, sizeBytes: 1_000, published: true, changed: false },
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

describe('hasChanges', () => {
  const published = new Set(['01 - Intro', '03 - Extra'])

  it('finds nothing to do while the selection matches what is on R2', () => {
    expect(hasChanges(plan(PUBLISHED_HERE), COURSE_ID, published)).toBe(false)
  })

  it('sees a chapter added or removed', () => {
    const added = new Set([...published, '02 - Deep'])
    expect(hasChanges(plan(PUBLISHED_HERE), COURSE_ID, added)).toBe(true)
    expect(hasChanges(plan(PUBLISHED_HERE), COURSE_ID, new Set(['01 - Intro']))).toBe(true)
  })

  it('sees changed files in a published chapter', () => {
    const edited = plan(PUBLISHED_HERE)
    const chapters = edited.chapters.map((chapter) => ({ ...chapter, changed: true }))
    expect(hasChanges({ ...edited, chapters }, COURSE_ID, published)).toBe(true)
  })

  it('treats a course not on R2 as new', () => {
    expect(hasChanges(plan(undefined), COURSE_ID, published)).toBe(true)
  })
})
