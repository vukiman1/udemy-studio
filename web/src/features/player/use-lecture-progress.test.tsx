import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render } from '@testing-library/react'
import { useRef, type ReactElement } from 'react'
import { afterEach, describe, expect, it, vi, type MockInstance } from 'vitest'
import type { LectureSummary } from '../../api/types'
import { useLectureProgress } from './use-lecture-progress'

const LECTURE: LectureSummary = {
  id: 'l1',
  number: 1,
  title: 'Welcome',
  kind: 'video',
  hasVideo: true,
  hasArticle: false,
  subtitleLanguages: [],
  attachments: [],
}

const SAVED_LECTURE: LectureSummary = {
  ...LECTURE,
  progress: {
    positionSeconds: 42,
    durationSeconds: 300,
    completed: false,
    updatedAt: '2026-11-24T08:00:00Z',
  },
}

interface HarnessProps {
  readonly lecture: LectureSummary
  readonly startAtSeconds: number | undefined
}

function Harness({ lecture, startAtSeconds }: HarnessProps): ReactElement {
  const videoRef = useRef<HTMLVideoElement | null>(null)
  useLectureProgress(videoRef, lecture, 'c1', startAtSeconds)
  return <video ref={videoRef} data-testid="video" />
}

function renderHarness(
  lecture: LectureSummary = LECTURE,
  startAtSeconds?: number,
): { video: HTMLVideoElement; unmount: () => void } {
  const client = new QueryClient()
  const { getByTestId, unmount } = render(
    <QueryClientProvider client={client}>
      <Harness lecture={lecture} startAtSeconds={startAtSeconds} />
    </QueryClientProvider>,
  )
  const video = getByTestId('video') as HTMLVideoElement
  Object.defineProperty(video, 'duration', { configurable: true, value: 300 })
  return { video, unmount }
}

function setPosition(video: HTMLVideoElement, seconds: number): void {
  Object.defineProperty(video, 'currentTime', {
    configurable: true,
    writable: true,
    value: seconds,
  })
}

async function savedBodies(fetchSpy: MockInstance<typeof fetch>): Promise<unknown[]> {
  const requests = fetchSpy.mock.calls
    .map(([input]) => input)
    .filter((input): input is Request => input instanceof Request)
  return Promise.all(requests.map((request): Promise<unknown> => request.clone().json()))
}

afterEach(() => {
  vi.restoreAllMocks()
})

describe('useLectureProgress', () => {
  it('starts at the requested second instead of where the lecture was left', () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(Response.json({}))
    const { video, unmount } = renderHarness(SAVED_LECTURE, 528)
    setPosition(video, 0)
    video.dispatchEvent(new Event('loadedmetadata'))
    expect(video.currentTime).toBe(528)
    unmount()
  })

  it('saves the position when the lecture is left mid-playback', async () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(Response.json({}))
    const { video, unmount } = renderHarness()
    setPosition(video, 298)
    unmount()
    await vi.waitFor(async () => {
      expect(await savedBodies(fetchSpy)).toContainEqual({
        positionSeconds: 298,
        durationSeconds: 300,
        completed: true,
      })
    })
  })

  it('saves a seek made while paused', async () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(Response.json({}))
    const { video } = renderHarness()
    setPosition(video, 120)
    video.dispatchEvent(new Event('seeked'))
    await vi.waitFor(async () => {
      expect(await savedBodies(fetchSpy)).toContainEqual({
        positionSeconds: 120,
        durationSeconds: 300,
        completed: false,
      })
    })
  })

  it('saves nothing for a lecture that was opened but never played', () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(Response.json({}))
    const { unmount } = renderHarness()
    unmount()
    expect(fetchSpy).not.toHaveBeenCalled()
  })
})
