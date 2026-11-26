import type { ReactElement } from 'react'
import { ButtonSize } from '../../ui/Button'
import { RemoveFromR2Button } from '../publish/RemoveFromR2Button'
import { useRemoveFromR2 } from '../publish/use-remove-from-r2'

export function RemovePublishedButton(): ReactElement {
  const removal = useRemoveFromR2()
  return (
    <span className="ml-auto flex items-center gap-3">
      {removal.isError && (
        <span className="text-danger" role="alert">
          {removal.error.message}
        </span>
      )}
      <RemoveFromR2Button removal={removal} label="Remove" size={ButtonSize.Small} />
    </span>
  )
}
