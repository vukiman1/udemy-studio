import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { ChevronLeft, FileQuestion } from 'lucide-react'
import { ButtonSize, ButtonVariant, buttonClass } from '../ui/Button'
import { EmptyState } from '../ui/EmptyState'
import { LIBRARY_PATH } from './paths'

export function NotFoundPage(): ReactElement {
  return (
    <main className="grid content-start px-5 py-8 md:px-10 md:py-10">
      <EmptyState
        icon={FileQuestion}
        title="Page not found."
        action={
          <Link to={LIBRARY_PATH} className={buttonClass(ButtonVariant.Primary, ButtonSize.Small)}>
            <ChevronLeft aria-hidden="true" />
            Back to library
          </Link>
        }
      />
    </main>
  )
}
