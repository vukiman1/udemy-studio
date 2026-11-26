import type { ReactElement } from 'react'

interface StaleDataNoticeProps {
  readonly error: Error
}

export function StaleDataNotice({ error }: StaleDataNoticeProps): ReactElement {
  return (
    <p className="m-0 px-4 py-1.5 text-[0.85rem] text-danger" role="status">
      Could not refresh ({error.message}); showing the last loaded data.
    </p>
  )
}
