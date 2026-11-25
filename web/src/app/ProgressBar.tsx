import type { ReactElement } from 'react'

interface ProgressBarProps {
  readonly completed: number
  readonly total: number
}

const PERCENT = 100

export function ProgressBar({ completed, total }: ProgressBarProps): ReactElement {
  const percent = total === 0 ? 0 : Math.round((completed / total) * PERCENT)
  return (
    <div
      className="h-1.5 overflow-hidden rounded-full bg-surface-raised"
      role="progressbar"
      aria-valuemin={0}
      aria-valuemax={PERCENT}
      aria-valuenow={percent}
    >
      <div className="h-full bg-accent" style={{ width: `${String(percent)}%` }} />
    </div>
  )
}
