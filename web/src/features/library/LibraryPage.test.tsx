import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import type { ReactElement } from 'react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { ContinueTarget, CourseSummary } from '../../api/types'
import { LibraryPage } from './LibraryPage'

const COURSE: CourseSummary = {
  id: 'c1',
  title: 'Backend Fundamentals',
  playableLectures: 4,
  completedLectures: 1,
}

function respondWith(routes: Record<string, () => Response>): void {
  vi.spyOn(globalThis, 'fetch').mockImplementation((input) => {
    const path = new URL(input instanceof Request ? input.url : String(input)).pathname
    const route = routes[path]
    return Promise.resolve(route === undefined ? new Response(null, { status: 404 }) : route())
  })
}

function renderPage(): void {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const tree: ReactElement = (
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <LibraryPage />
      </MemoryRouter>
    </QueryClientProvider>
  )
  render(tree)
}

afterEach(() => {
  vi.restoreAllMocks()
})

describe('LibraryPage', () => {
  it('lists courses with their completion', async () => {
    respondWith({
      '/api/courses': () => Response.json([COURSE]),
      '/api/continue': () => new Response(null, { status: 204 }),
    })
    renderPage()
    expect(await screen.findByText('Backend Fundamentals')).toBeInTheDocument()
    expect(screen.getByText('1/4 bài đã học')).toBeInTheDocument()
    expect(screen.queryByText(/Tiếp tục học/)).not.toBeInTheDocument()
  })

  it('offers to continue where the learner left off', async () => {
    const target: ContinueTarget = { courseId: 'c1', lectureId: 'l7', positionSeconds: 12 }
    respondWith({
      '/api/courses': () => Response.json([COURSE]),
      '/api/continue': () => Response.json(target),
    })
    renderPage()
    const link = await screen.findByRole('link', { name: /Tiếp tục học/ })
    expect(link).toHaveAttribute('href', '/courses/c1/lectures/l7')
  })

  it('explains when the library cannot be loaded', async () => {
    respondWith({ '/api/continue': () => new Response(null, { status: 204 }) })
    renderPage()
    expect(await screen.findByRole('alert')).toHaveTextContent('Không tải được thư viện')
  })

  it('says so when the downloads folder is empty', async () => {
    respondWith({
      '/api/courses': () => Response.json([]),
      '/api/continue': () => new Response(null, { status: 204 }),
    })
    renderPage()
    expect(await screen.findByText(/Chưa có khoá học nào/)).toBeInTheDocument()
  })
})

describe('LibraryPage after a failed background refresh', () => {
  it('keeps showing the courses it already has', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    let failNext = false
    vi.spyOn(globalThis, 'fetch').mockImplementation((input) => {
      const path = new URL(input instanceof Request ? input.url : String(input)).pathname
      if (path === '/api/continue') return Promise.resolve(new Response(null, { status: 204 }))
      return Promise.resolve(
        failNext ? new Response(null, { status: 503 }) : Response.json([COURSE]),
      )
    })
    render(
      <QueryClientProvider client={client}>
        <MemoryRouter>
          <LibraryPage />
        </MemoryRouter>
      </QueryClientProvider>,
    )
    expect(await screen.findByText('Backend Fundamentals')).toBeInTheDocument()
    failNext = true
    await client.refetchQueries({ queryKey: ['courses'] })
    expect(await screen.findByText(/Không làm mới được/)).toBeInTheDocument()
    expect(screen.getByText('Backend Fundamentals')).toBeInTheDocument()
  })
})
