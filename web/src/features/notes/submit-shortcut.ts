import type { KeyboardEvent } from 'react'

export function isSubmitShortcut(event: KeyboardEvent): boolean {
  return event.key === 'Enter' && (event.metaKey || event.ctrlKey)
}
