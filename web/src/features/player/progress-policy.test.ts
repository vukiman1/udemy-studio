import { describe, expect, it } from 'vitest'
import type { LectureProgress } from '../../api/types'
import { isSaveDue, isWatchedToCompletion, resumePosition } from './progress-policy'

function saved(positionSeconds: number, completed: boolean): LectureProgress {
  return { positionSeconds, durationSeconds: 100, completed, updatedAt: '2026-11-23T08:00:00Z' }
}

describe('isWatchedToCompletion', () => {
  it('counts 95% as finished', () => {
    expect(isWatchedToCompletion(95, 100)).toBe(true)
  })
  it('does not count 94%', () => {
    expect(isWatchedToCompletion(94, 100)).toBe(false)
  })
  it('never completes without a known duration', () => {
    expect(isWatchedToCompletion(10, 0)).toBe(false)
  })
})

describe('resumePosition', () => {
  it('rewinds three seconds to rejoin the sentence', () => {
    expect(resumePosition(saved(42, false))).toBe(39)
  })
  it('never goes before the start', () => {
    expect(resumePosition(saved(1, false))).toBe(0)
  })
  it('restarts finished lectures from the beginning', () => {
    expect(resumePosition(saved(98, true))).toBe(0)
  })
  it('starts fresh lectures at zero', () => {
    expect(resumePosition(undefined)).toBe(0)
  })
})

describe('isSaveDue', () => {
  it('waits five seconds between saves', () => {
    expect(isSaveDue(10_000, 14_999)).toBe(false)
    expect(isSaveDue(10_000, 15_000)).toBe(true)
  })
})
