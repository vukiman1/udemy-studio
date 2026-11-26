import type { ReactElement } from 'react'
import { ChevronLeft, FileQuestion } from 'lucide-react'
import { Button, ButtonSize, ButtonVariant } from '../ui/Button'
import { EmptyState } from '../ui/EmptyState'
import { LIBRARY_PATH } from './paths'

export function NotFoundPage(): ReactElement {
  return (
    <main className="grid content-start px-5 py-8 md:px-10 md:py-10">
      <EmptyState
        icon={FileQuestion}
        title="Page not found."
        action={
          <Button
            to={LIBRARY_PATH}
            variant={ButtonVariant.Primary}
            size={ButtonSize.Small}
            icon={ChevronLeft}
          >
            Back to library
          </Button>
        }
      />
    </main>
  )
}
