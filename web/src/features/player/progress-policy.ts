import type { LectureProgress } from '../../api/types'

export const PROGRESS_SAVE_INTERVAL_MS = 5000
export const COMPLETION_RATIO = 0.95
export const RESUME_REWIND_SECONDS = 3

export function isWatchedToCompletion(positionSeconds: number, durationSeconds: number): boolean {
  return durationSeconds > 0 && positionSeconds / durationSeconds >= COMPLETION_RATIO
}

export function resumePosition(saved: LectureProgress | undefined): number {
  if (saved === undefined || saved.completed) return 0
  return Math.max(0, saved.positionSeconds - RESUME_REWIND_SECONDS)
}

export function isSaveDue(lastSavedAtMs: number, nowMs: number): boolean {
  return nowMs - lastSavedAtMs >= PROGRESS_SAVE_INTERVAL_MS
}
