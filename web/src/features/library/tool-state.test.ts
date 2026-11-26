import { describe, expect, it } from 'vitest'
import type { CourseTools, PublishRun, TranslationRun } from '../../api/types'
import { ToolState, compressionStatus, publishStatus, translationStatus } from './tool-state'

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
    expect(status.label).toBe('Translating 4/10')
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

function publishRun(state: PublishRun['state']): PublishRun {
  return {
    courseId: 'course',
    title: 'Course',
    state,
    totalFiles: 40,
    sentFiles: 10,
    skippedFiles: 10,
    totalBytes: 1000,
    sentBytes: 250,
  }
}

describe('publishStatus', () => {
  it('follows the bytes sent while a publish runs', () => {
    const status = publishStatus(publishRun('running'), false)
    expect(status.state).toBe(ToolState.Running)
    expect(status.progress).toBeCloseTo(0.25)
    expect(status.label).toBe('Publishing 25%')
  })

  it('stays open on a published course so changed files can be sent again', () => {
    const status = publishStatus(publishRun('done'), true)
    expect(status.state).toBe(ToolState.Idle)
    expect(status.label).toBe('Update R2')
  })

  it('shows a failed publish until it is retried', () => {
    expect(publishStatus(publishRun('failed'), false).state).toBe(ToolState.Failed)
  })
})
