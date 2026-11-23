import { useEffect, useState, type RefObject } from 'react'
import type { Cue, CueLine } from '../../api/types'
import { activeLinesAt } from './active-cues-at'

const MILLISECONDS_PER_SECOND = 1000
const NO_LINES: readonly CueLine[] = []

export function useActiveLines(
  videoRef: RefObject<HTMLVideoElement | null>,
  cues: readonly Cue[],
  languageOrder: readonly string[],
): readonly CueLine[] {
  const [lines, setLines] = useState<readonly CueLine[]>(NO_LINES)
  const languageKey = languageOrder.join(',')

  useEffect(() => {
    const video = videoRef.current
    if (video === null) return
    const order = languageKey.split(',')
    let frame = 0
    const sync = (): void => {
      const next = activeLinesAt(cues, video.currentTime * MILLISECONDS_PER_SECOND, order)
      setLines((previous) => (sameLines(previous, next) ? previous : next))
    }
    const followPlayback = (): void => {
      sync()
      frame = window.requestAnimationFrame(followPlayback)
    }
    const startFollowing = (): void => {
      window.cancelAnimationFrame(frame)
      followPlayback()
    }
    const stopFollowing = (): void => {
      window.cancelAnimationFrame(frame)
      sync()
    }
    sync()
    if (!video.paused) startFollowing()
    video.addEventListener('play', startFollowing)
    video.addEventListener('pause', stopFollowing)
    video.addEventListener('ended', stopFollowing)
    video.addEventListener('seeked', sync)
    return () => {
      window.cancelAnimationFrame(frame)
      video.removeEventListener('play', startFollowing)
      video.removeEventListener('pause', stopFollowing)
      video.removeEventListener('ended', stopFollowing)
      video.removeEventListener('seeked', sync)
    }
  }, [videoRef, cues, languageKey])

  return lines
}

function sameLines(previous: readonly CueLine[], next: readonly CueLine[]): boolean {
  return previous.length === next.length && previous.every((line, index) => line === next[index])
}
