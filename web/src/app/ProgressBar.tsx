import type { ReactElement } from 'react'
import styles from './ProgressBar.module.css'

interface ProgressBarProps {
  readonly completed: number
  readonly total: number
}

const PERCENT = 100

export function ProgressBar({ completed, total }: ProgressBarProps): ReactElement {
  const percent = total === 0 ? 0 : Math.round((completed / total) * PERCENT)
  return (
    <div
      className={styles.track}
      role="progressbar"
      aria-valuemin={0}
      aria-valuemax={PERCENT}
      aria-valuenow={percent}
    >
      <div className={styles.fill} style={{ width: `${String(percent)}%` }} />
    </div>
  )
}
