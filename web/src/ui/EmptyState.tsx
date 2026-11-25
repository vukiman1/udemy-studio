import type { LucideIcon } from 'lucide-react'
import type { ReactElement, ReactNode } from 'react'

interface EmptyStateProps {
  readonly icon: LucideIcon
  readonly title: string
  readonly body?: string
  readonly action?: ReactNode
  // An error is announced to screen readers as soon as it appears; anything else waits its turn.
  readonly isError?: boolean
}

export function EmptyState({
  icon: Icon,
  title,
  body,
  action,
  isError = false,
}: EmptyStateProps): ReactElement {
  return (
    <section
      className="grid justify-items-center gap-2.5 rounded-2xl bg-card px-8 py-16 text-center"
      role={isError ? 'alert' : 'status'}
    >
      <Icon className="size-9 text-muted" strokeWidth={1.6} aria-hidden="true" />
      <h2 className="m-0 text-lg font-semibold">{title}</h2>
      {body !== undefined && <p className="m-0 text-sm text-muted">{body}</p>}
      {action !== undefined && <div className="mt-1">{action}</div>}
    </section>
  )
}
