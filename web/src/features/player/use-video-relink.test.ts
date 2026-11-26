import { renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it, vi, type MockInstance } from 'vitest'
import { MediaErrorCode, useVideoRelink } from './use-video-relink'

const POSITION_SECONDS = 1234
const RELINK_RETRY_AFTER_MS = 31_000

interface FailingVideo {
  readonly video: HTMLVideoElement
  readonly load: MockInstance
  readonly play: MockInstance
}

function videoThatFails(code: MediaErrorCode): FailingVideo {
  const video = document.createElement('video')
  let currentTime = POSITION_SECONDS
  Object.defineProperty(video, 'error', { value: { code } })
  Object.defineProperty(video, 'currentTime', {
    get: () => currentTime,
    set: (seconds: number) => {
      currentTime = seconds
    },
  })
  const load = vi.spyOn(video, 'load').mockImplementation(() => {
    currentTime = 0
  })
  const play = vi.spyOn(video, 'play').mockResolvedValue(undefined)
  return { video, load, play }
}

function failAndReload(video: HTMLVideoElement, isPlaying: boolean): void {
  renderHook(() => {
    useVideoRelink({ current: video })
  })
  if (isPlaying) video.dispatchEvent(new Event('play'))
  video.dispatchEvent(new Event('error'))
  video.dispatchEvent(new Event('loadedmetadata'))
}

describe('useVideoRelink', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('reloads an expired link and plays on from the same moment, though the browser already paused', () => {
    const { video, load, play } = videoThatFails(MediaErrorCode.Network)
    failAndReload(video, true)
    expect(load).toHaveBeenCalledTimes(1)
    expect(video.currentTime).toBe(POSITION_SECONDS)
    expect(play).toHaveBeenCalledTimes(1)
  })

  it('brings a paused video back paused at the same moment', () => {
    const { video, load, play } = videoThatFails(MediaErrorCode.Network)
    failAndReload(video, false)
    expect(load).toHaveBeenCalledTimes(1)
    expect(video.currentTime).toBe(POSITION_SECONDS)
    expect(play).not.toHaveBeenCalled()
  })

  it('tries once and then leaves a video that keeps failing alone', () => {
    const { video, load } = videoThatFails(MediaErrorCode.Network)
    failAndReload(video, true)
    video.dispatchEvent(new Event('error'))
    expect(load).toHaveBeenCalledTimes(1)
  })

  it('treats a replay the browser refuses as paused', async () => {
    vi.useFakeTimers()
    const { video, load, play } = videoThatFails(MediaErrorCode.Network)
    play.mockRejectedValueOnce(new DOMException('needs a tap', 'NotAllowedError'))
    failAndReload(video, true)
    await vi.advanceTimersByTimeAsync(RELINK_RETRY_AFTER_MS)
    video.dispatchEvent(new Event('error'))
    video.dispatchEvent(new Event('loadedmetadata'))
    expect(load).toHaveBeenCalledTimes(2)
    expect(play).toHaveBeenCalledTimes(1)
    vi.useRealTimers()
  })

  it('does not reload a video that cannot be decoded', () => {
    const { video, load } = videoThatFails(MediaErrorCode.Decode)
    failAndReload(video, true)
    expect(load).not.toHaveBeenCalled()
  })
})
