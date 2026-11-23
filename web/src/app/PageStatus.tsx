import type { ReactElement } from 'react'

interface PageStatusProps {
  readonly message: string
  readonly tone: 'info' | 'error'
}

export function PageStatus({ message, tone }: PageStatusProps): ReactElement {
  return (
    <p
      className={tone === 'error' ? 'm-0 p-6 text-danger' : 'm-0 p-6'}
      role={tone === 'error' ? 'alert' : 'status'}
    >
      {message}
    </p>
  )
}
