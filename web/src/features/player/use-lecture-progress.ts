import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useEffectEvent, useState, type RefObject } from 'react'
import { QUERY_KEYS, saveProgress } from '../../api/queries'
import type { LectureSummary } from '../../api/types'
import { isSaveDue, isWatchedToCompletion, resumePosition } from './progress-policy'

interface ProgressSaveState {
  readonly lastError: Error | undefined
}

interface WatchSession {
  lastSavedAtMs: number
  lastSavedPosition: number
  isCompleted: boolean
}

export function useLectureProgress(
  videoRef: RefObject<HTMLVideoElement | null>,
  lecture: LectureSummary,
  courseId: string,
  startAtSeconds: number | undefined,
): ProgressSaveState {
  const queryClient = useQueryClient()
  const [lastError, setLastError] = useState<Error | undefined>(undefined)
  const lectureId = lecture.id

  const savedProgress = useEffectEvent(() => lecture.progress)
  const startPosition = useEffectEvent(() => startAtSeconds ?? resumePosition(lecture.progress))
  const onSaved = useEffectEvent((isNewlyCompleted: boolean) => {
    setLastError(undefined)
    if (!isNewlyCompleted) return
    void queryClient.invalidateQueries({ queryKey: QUERY_KEYS.course(courseId) })
    void queryClient.invalidateQueries({ queryKey: QUERY_KEYS.courses })
    void queryClient.invalidateQueries({ queryKey: QUERY_KEYS.continueTarget })
  })
  const onSaveFailed = useEffectEvent((error: unknown) => {
    setLastError(error instanceof Error ? error : new Error(String(error)))
  })

  useEffect(() => {
    const video = videoRef.current
    if (video === null) return
    const session: WatchSession = {
      lastSavedAtMs: Date.now(),
      lastSavedPosition: 0,
      isCompleted: savedProgress()?.completed ?? false,
    }

    const restorePosition = (): void => {
      video.currentTime = startPosition()
    }
    const report = (shouldKeepAlive: boolean, hasEnded: boolean): void => {
      const duration = video.duration
      if (!Number.isFinite(duration) || duration <= 0) return
      const position = hasEnded ? duration : video.currentTime
      const completed = hasEnded || isWatchedToCompletion(position, duration)
      const isNewlyCompleted = completed && !session.isCompleted
      session.isCompleted ||= completed
      session.lastSavedAtMs = Date.now()
      session.lastSavedPosition = position
      saveProgress(
        lectureId,
        { positionSeconds: position, durationSeconds: duration, completed },
        shouldKeepAlive,
      ).then(() => {
        onSaved(isNewlyCompleted)
      }, onSaveFailed)
    }
    const onTimeUpdate = (): void => {
      if (!video.paused && isSaveDue(session.lastSavedAtMs, Date.now())) report(false, false)
    }
    const onPause = (): void => {
      if (!video.ended) report(false, false)
    }
    const onEnded = (): void => {
      report(false, true)
    }
    const onPageHidden = (): void => {
      if (document.visibilityState === 'hidden') report(true, false)
    }
    const onSeekedWhilePaused = (): void => {
      if (video.paused) report(false, false)
    }
    const saveUnsavedPosition = (): void => {
      if (video.currentTime > 0 && video.currentTime !== session.lastSavedPosition)
        report(true, false)
    }

    if (video.readyState >= HTMLMediaElement.HAVE_METADATA) restorePosition()
    video.addEventListener('loadedmetadata', restorePosition)
    video.addEventListener('timeupdate', onTimeUpdate)
    video.addEventListener('pause', onPause)
    video.addEventListener('ended', onEnded)
    video.addEventListener('seeked', onSeekedWhilePaused)
    document.addEventListener('visibilitychange', onPageHidden)
    window.addEventListener('pagehide', onPageHidden)
    return () => {
      video.removeEventListener('loadedmetadata', restorePosition)
      video.removeEventListener('timeupdate', onTimeUpdate)
      video.removeEventListener('pause', onPause)
      video.removeEventListener('ended', onEnded)
      video.removeEventListener('seeked', onSeekedWhilePaused)
      document.removeEventListener('visibilitychange', onPageHidden)
      window.removeEventListener('pagehide', onPageHidden)
      saveUnsavedPosition()
    }
  }, [videoRef, lectureId])

  return { lastError }
}
