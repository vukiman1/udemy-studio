import { CircleAlert, LoaderCircle, type LucideIcon } from 'lucide-react'
import type { ReactElement } from 'react'
import { Button, ButtonVariant } from '../../ui/Button'
import { ToolState, type ToolStatus } from './tool-state'

interface CourseToolButtonProps {
  readonly icon: LucideIcon
  readonly status: ToolStatus
  readonly onOpen: () => void
}

const STATE_VARIANTS: Readonly<Record<ToolState, ButtonVariant>> = {
  [ToolState.Idle]: ButtonVariant.Secondary,
  [ToolState.Running]: ButtonVariant.Secondary,
  [ToolState.Done]: ButtonVariant.Secondary,
  [ToolState.Failed]: ButtonVariant.Danger,
  [ToolState.Checking]: ButtonVariant.Secondary,
  [ToolState.Unavailable]: ButtonVariant.Secondary,
}

const CLOSED_STATES: ReadonlySet<ToolState> = new Set([
  ToolState.Done,
  ToolState.Checking,
  ToolState.Unavailable,
])

function stateIcon(status: ToolStatus, icon: LucideIcon): LucideIcon {
  if (status.state === ToolState.Running || status.state === ToolState.Checking) return LoaderCircle
  if (status.state === ToolState.Failed) return CircleAlert
  return icon
}

export function CourseToolButton({ icon, status, onOpen }: CourseToolButtonProps): ReactElement {
  const isBusy = status.state === ToolState.Running || status.state === ToolState.Checking
  const isDisabled = CLOSED_STATES.has(status.state)
  const button = (
    <Button
      variant={STATE_VARIANTS[status.state]}
      icon={stateIcon(status, icon)}
      className={isBusy ? '[&_svg]:animate-spin' : ''}
      title={isDisabled ? undefined : status.detail}
      aria-label={`${status.label}: ${status.detail}`}
      disabled={isDisabled}
      onClick={onOpen}
    >
      {status.label}
    </Button>
  )
  if (!isDisabled) return button
  return (
    <span className="group/tip relative inline-flex">
      {button}
      <span
        aria-hidden="true"
        className="pointer-events-none absolute right-0 bottom-full z-10 mb-2 rounded-lg bg-ink px-2.5 py-1.5 text-xs whitespace-nowrap text-page opacity-0 transition-opacity group-hover/tip:opacity-100"
      >
        {status.detail}
      </span>
    </span>
  )
}
