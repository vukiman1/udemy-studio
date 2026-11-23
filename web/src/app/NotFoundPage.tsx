import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { PageStatus } from './PageStatus'

export function NotFoundPage(): ReactElement {
  return (
    <main>
      <PageStatus tone="error" message="Không có trang này." />
      <Link className="px-6 text-accent underline" to="/">
        ← Về thư viện
      </Link>
    </main>
  )
}
