import type { ReactElement } from 'react'
import { Button, ButtonVariant } from '../ui/Button'

interface RunInBackgroundButtonProps {
  readonly onClick: () => void
}

// Closing the dialog never stops a run; this button says so, for learners who want to keep watching.
export function RunInBackgroundButton({ onClick }: RunInBackgroundButtonProps): ReactElement {
  return (
    <Button variant={ButtonVariant.Primary} onClick={onClick}>
      Run in background
    </Button>
  )
}
