import { useEffect, type RefObject } from 'react'

const NATIVE_FULLSCREEN_BEGIN = 'webkitbeginfullscreen'
const NATIVE_FULLSCREEN_END = 'webkitendfullscreen'

// iPhone fullscreen hides our overlay, so a <track> is attached only for that span.
export function useNativeFullscreenSubtitles(
  videoRef: RefObject<HTMLVideoElement | null>,
  trackSource: string | undefined,
  language: string | undefined,
): void {
  useEffect(() => {
    const video = videoRef.current
    if (video === null || trackSource === undefined) return
    let track: HTMLTrackElement | undefined

    const attach = (): void => {
      track = document.createElement('track')
      track.kind = 'subtitles'
      track.src = trackSource
      track.srclang = language ?? ''
      track.label = 'Phụ đề'
      track.default = true
      video.appendChild(track)
      track.track.mode = 'showing'
    }
    const detach = (): void => {
      track?.remove()
      track = undefined
    }

    video.addEventListener(NATIVE_FULLSCREEN_BEGIN, attach)
    video.addEventListener(NATIVE_FULLSCREEN_END, detach)
    return () => {
      video.removeEventListener(NATIVE_FULLSCREEN_BEGIN, attach)
      video.removeEventListener(NATIVE_FULLSCREEN_END, detach)
      detach()
    }
  }, [videoRef, trackSource, language])
}
