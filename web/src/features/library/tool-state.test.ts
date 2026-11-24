import { describe, expect, it } from 'vitest'
import type { CourseTools, TranslationRun } from '../../api/types'
import { ToolState, compressionStatus, translationStatus } from './tool-state'

function translationRun(state: TranslationRun['state'], failed = 0): TranslationRun {
  return {
    state,
    total: 10,
    finished: 4,
    costUsd: 0,
    failed: Array.from({ length: failed }, (_, position) => ({
      path: `p${String(position)}`,
      lectureTitle: 'L',
      reason: 'r',
    })),
  }
}

type Translation = CourseTools['translation']

describe('translationStatus', () => {
  it('reports progress while a run is going', () => {
    const status = translationStatus({ lectures: 10, pending: 6, run: translationRun('running') })
    expect(status.state).toBe(ToolState.Running)
    expect(status.progress).toBeCloseTo(0.4)
  })

  it('is done once every lecture has Vietnamese subtitles, whatever the last run did', () => {
    expect(
      translationStatus({ lectures: 10, pending: 0, run: translationRun('done', 2) }).state,
    ).toBe(ToolState.Done)
  })

  it('opens again when lectures without Vietnamese subtitles are added after a finished run', () => {
    const afterNewDownload: Translation = { lectures: 12, pending: 2, run: translationRun('done') }
    expect(translationStatus(afterNewDownload).state).toBe(ToolState.Idle)
  })

  it('shows a failure while lectures are left after a run that had errors', () => {
    expect(
      translationStatus({ lectures: 10, pending: 3, run: translationRun('done', 3) }).state,
    ).toBe(ToolState.Failed)
  })
})

describe('compressionStatus', () => {
  it('waits for the survey before calling a course done', () => {
    expect(compressionStatus({ videos: 5, pending: 0, unchecked: 2 }).state).toBe(
      ToolState.Checking,
    )
    expect(compressionStatus({ videos: 5, pending: 0, unchecked: 0 }).state).toBe(ToolState.Done)
  })

  it('is unavailable without ffmpeg', () => {
    expect(compressionStatus(undefined).state).toBe(ToolState.Unavailable)
  })
})
