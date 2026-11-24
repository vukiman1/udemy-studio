import { LoaderCircle, type LucideIcon } from 'lucide-react'
import type { ReactElement } from 'react'
import { ToolState, type ToolStatus } from './tool-state'

interface CourseToolButtonProps {
  readonly icon: LucideIcon
  readonly status: ToolStatus
  readonly onOpen: () => void
}

const STATE_STYLES: Readonly<Record<ToolState, string>> = {
  [ToolState.Idle]: 'text-muted hover:border-accent hover:text-accent',
  [ToolState.Running]: 'border-accent text-accent',
  // A finished tool stays fully coloured: disabled because there is nothing left, not because it is unavailable.
  [ToolState.Done]: 'border-success/50 text-success disabled:opacity-100',
  [ToolState.Failed]: 'border-danger/60 text-danger',
  [ToolState.Checking]: 'text-muted',
  [ToolState.Unavailable]: 'text-muted',
}

const CLOSED_STATES: ReadonlySet<ToolState> = new Set([
  ToolState.Done,
  ToolState.Checking,
  ToolState.Unavailable,
])

export function CourseToolButton({
  icon: Icon,
  status,
  onOpen,
}: CourseToolButtonProps): ReactElement {
  return (
    <button
      type="button"
      className={`relative inline-flex size-9 items-center justify-center rounded-lg p-0 ${STATE_STYLES[status.state]}`}
      title={status.detail}
      aria-label={status.detail}
      disabled={CLOSED_STATES.has(status.state)}
      onClick={onOpen}
    >
      <Icon className="size-4.5" />
      {status.state === ToolState.Running && (
        <LoaderCircle className="absolute -top-1 -right-1 size-3.5 animate-spin rounded-full bg-surface" />
      )}
    </button>
  )
}
