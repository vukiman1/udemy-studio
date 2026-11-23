import type { ReactElement } from 'react'
import styles from './PageStatus.module.css'

interface PageStatusProps {
  readonly message: string
  readonly tone: 'info' | 'error'
}

export function PageStatus({ message, tone }: PageStatusProps): ReactElement {
  return (
    <p
      className={tone === 'error' ? styles.error : styles.info}
      role={tone === 'error' ? 'alert' : 'status'}
    >
      {message}
    </p>
  )
}
