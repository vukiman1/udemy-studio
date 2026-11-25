import type { ReactElement } from 'react'

interface ProgressBarProps {
  readonly completed: number
  readonly total: number
  readonly className?: string
}

const PERCENT = 100

export function ProgressBar({
  completed,
  total,
  className = 'h-1.5 rounded-full',
}: ProgressBarProps): ReactElement {
  const percent = completionPercent(completed, total)
  return (
    <div
      className={`overflow-hidden bg-track ${className}`}
      role="progressbar"
      aria-valuemin={0}
      aria-valuemax={PERCENT}
      aria-valuenow={percent}
    >
      <div
        className="h-full bg-accent transition-[width] duration-700 ease-soft"
        style={{ width: `${String(percent)}%` }}
      />
    </div>
  )
}

export function completionPercent(completed: number, total: number): number {
  return total === 0 ? 0 : Math.round((completed / total) * PERCENT)
}
