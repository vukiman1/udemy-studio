import { PlayButton } from '@videojs/react'
import { LoaderCircle, Play } from 'lucide-react'
import type { ReactElement } from 'react'
import { usePlayer } from './player'
import { useDelayedFlag } from './use-delayed-flag'

const LOADING_DELAY_MS = 400

export function PlayerCenterState(): ReactElement | null {
  const isLoading = usePlayer(
    (state) => state.error === null && (state.duration === 0 || (state.waiting && !state.paused)),
  )
  const isPaused = usePlayer((state) => state.paused)
  const isSpinnerShown = useDelayedFlag(isLoading, LOADING_DELAY_MS)

  if (isSpinnerShown)
    return (
      <div
        className="pointer-events-none absolute inset-0 grid place-items-center"
        role="status"
        aria-label="Đang tải video"
      >
        <span className="grid size-16 place-items-center rounded-full bg-black/45 backdrop-blur-sm">
          <LoaderCircle className="size-8 animate-spin" aria-hidden="true" />
        </span>
      </div>
    )
  if (!isPaused || isLoading) return null
  return (
    <PlayButton
      className="absolute top-1/2 left-1/2 grid size-16 -translate-1/2 place-items-center rounded-full bg-white/12 p-0 text-white backdrop-blur-sm hover:bg-white/20"
      render={(attributes) => (
        <button {...attributes} tabIndex={-1} aria-hidden="true">
          <Play className="ml-1 size-7 fill-current" aria-hidden="true" />
        </button>
      )}
    />
  )
}
