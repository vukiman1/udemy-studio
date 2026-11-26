import { describe, expect, it } from 'vitest'
import type { Job } from '../../api/types'
import { jobErrorText } from './job-error'

const FAILED_JOB: Job = {
  id: 'j1',
  courseUrl: 'https://acme.udemy.com/course/x/',
  status: 'failed',
  createdAt: '2026-11-25T08:00:00Z',
  progress: { courseTitle: 'X', totalLectures: 0, finishedCount: 0 },
  failures: [],
  error: 'udemy returned HTTP 401 for https://acme.udemy.com/api-2.0/courses/1/',
}

describe('jobErrorText', () => {
  it('tells what to do when the server knows why the download stopped', () => {
    expect(jobErrorText({ ...FAILED_JOB, errorKind: 'login' })).toBe(
      'Udemy login expired. Paste a new token and download again.',
    )
  })

  it('shows the server message when the error has no known cause', () => {
    expect(jobErrorText({ ...FAILED_JOB, errorKind: 'other' })).toBe(FAILED_JOB.error)
    expect(jobErrorText({ ...FAILED_JOB, errorKind: undefined })).toBe(FAILED_JOB.error)
  })
})
