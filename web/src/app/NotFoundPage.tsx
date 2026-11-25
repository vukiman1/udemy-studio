import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { PageStatus } from './PageStatus'
import { ChevronIcon } from '@videojs/react/icons'
import { LIBRARY_PATH } from './paths'

export function NotFoundPage(): ReactElement {
  return (
    <main>
      <PageStatus tone="error" message="Không có trang này." />
      <Link className="inline-flex items-center gap-1 px-6 text-accent underline" to={LIBRARY_PATH}>
        <ChevronIcon className="size-4.5 -scale-x-100" />
        Về thư viện
      </Link>
    </main>
  )
}
