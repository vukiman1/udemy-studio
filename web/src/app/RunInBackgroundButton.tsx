import type { ReactElement } from 'react'

interface RunInBackgroundButtonProps {
  readonly onClick: () => void
}

// Closing the dialog never stops a run; this button says so, for learners who want to keep watching.
export function RunInBackgroundButton({ onClick }: RunInBackgroundButtonProps): ReactElement {
  return (
    <button
      type="button"
      className="border-accent bg-accent font-semibold text-on-accent"
      onClick={onClick}
    >
      Chạy nền
    </button>
  )
}
