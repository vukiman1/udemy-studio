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
  [ToolState.Done]: ButtonVariant.Done,
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
  return (
    <Button
      variant={STATE_VARIANTS[status.state]}
      icon={stateIcon(status, icon)}
      className={isBusy ? '[&_svg]:animate-spin' : ''}
      title={status.detail}
      aria-label={`${status.label}: ${status.detail}`}
      disabled={CLOSED_STATES.has(status.state)}
      onClick={onOpen}
    >
      {status.label}
    </Button>
  )
}
