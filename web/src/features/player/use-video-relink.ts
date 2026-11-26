import { useEffect, type RefObject } from 'react'

export enum MediaErrorCode {
  Network = 2,
  Decode = 3,
  SourceNotSupported = 4,
}

const RELINK_COOLDOWN_MS = 30_000
const RELINKABLE_ERRORS: ReadonlySet<number> = new Set([
  MediaErrorCode.Network,
  MediaErrorCode.SourceNotSupported,
])

export function useVideoRelink(videoRef: RefObject<HTMLVideoElement | null>): void {
  useEffect(() => {
    const video = videoRef.current
    if (video === null) return
    let lastRelinkAtMs = Number.NEGATIVE_INFINITY
    let isMeantToPlay = !video.paused
    const markPlaying = (): void => {
      isMeantToPlay = true
    }
    const markStopped = (): void => {
      isMeantToPlay = false
    }

    const relink = (): void => {
      const code = video.error?.code
      if (code === undefined || !RELINKABLE_ERRORS.has(code)) return
      if (Date.now() - lastRelinkAtMs < RELINK_COOLDOWN_MS) return
      lastRelinkAtMs = Date.now()
      const position = video.currentTime
      const wasPlaying = isMeantToPlay
      video.addEventListener(
        'loadedmetadata',
        () => {
          video.currentTime = position
          if (wasPlaying) video.play().catch(markStopped)
        },
        { once: true },
      )
      video.load()
    }

    video.addEventListener('error', relink)
    video.addEventListener('play', markPlaying)
    video.addEventListener('pause', markStopped)
    video.addEventListener('ended', markStopped)
    return () => {
      video.removeEventListener('error', relink)
      video.removeEventListener('play', markPlaying)
      video.removeEventListener('pause', markStopped)
      video.removeEventListener('ended', markStopped)
    }
  }, [videoRef])
}
