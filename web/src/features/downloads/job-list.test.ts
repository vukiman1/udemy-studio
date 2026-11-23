import { describe, expect, it } from 'vitest'
import type { Job } from '../../api/types'
import { isActive, parseJobEvent, upsertJob } from './job-list'

function job(id: string, createdAt: string, overrides: Partial<Job> = {}): Job {
  return {
    id,
    courseUrl: 'https://acme.udemy.com/course/x/',
    status: 'queued',
    createdAt,
    failures: [],
    progress: { courseTitle: '', totalLectures: 0, finishedCount: 0 },
    ...overrides,
  }
}

describe('upsertJob', () => {
  it('replaces the job with the same id', () => {
    const before = [job('a', '2026-11-23T08:00:00Z')]
    const updated = upsertJob(before, job('a', '2026-11-23T08:00:00Z', { status: 'running' }))
    expect(updated).toHaveLength(1)
    expect(updated[0]?.status).toBe('running')
  })

  it('adds a new job and keeps the newest first', () => {
    const before = [job('old', '2026-11-23T08:00:00Z')]
    const updated = upsertJob(before, job('new', '2026-11-23T09:00:00Z'))
    expect(updated.map((item) => item.id)).toEqual(['new', 'old'])
  })
})

describe('isActive', () => {
  it('treats queued and running jobs as active', () => {
    expect(isActive(job('a', 'x'))).toBe(true)
    expect(isActive(job('a', 'x', { status: 'running' }))).toBe(true)
    expect(isActive(job('a', 'x', { status: 'cancelled' }))).toBe(false)
  })
})

describe('parseJobEvent', () => {
  it('accepts a job sent by the server', () => {
    const sent = job('a', '2026-11-23T08:00:00Z', { status: 'running' })
    expect(parseJobEvent(JSON.stringify(sent))).toEqual(sent)
  })

  it.each([
    'not json',
    '42',
    '{"id":"a"}',
    JSON.stringify({ ...job('a', 'x'), status: 'exploded' }),
  ])('rejects %s', (data) => {
    expect(parseJobEvent(data)).toBeUndefined()
  })
})
