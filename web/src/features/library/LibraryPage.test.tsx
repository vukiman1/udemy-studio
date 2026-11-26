import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen } from '@testing-library/react'
import type { ReactElement } from 'react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { CourseSummary } from '../../api/types'
import { LibraryPage } from './LibraryPage'

const COURSE: CourseSummary = {
  id: 'c1',
  title: 'Backend Fundamentals',
  playableLectures: 4,
  completedLectures: 1,
  sizeBytes: 15_800_000_000,
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
  it('lists courses with their completion and size', async () => {
    respondWith({ '/api/courses': () => Response.json([COURSE]) })
    renderPage()
    expect(await screen.findByText('Backend Fundamentals')).toBeInTheDocument()
    expect(screen.getByText('1 / 4 lectures · 15.8 GB')).toBeInTheDocument()
  })

  it('continues each course where the learner left it', async () => {
    const resumed: CourseSummary = {
      ...COURSE,
      id: 'c2',
      title: 'Resumed',
      resume: { courseId: 'c2', lectureId: 'l7', positionSeconds: 12 },
    }
    respondWith({ '/api/courses': () => Response.json([COURSE, resumed]) })
    renderPage()
    const links = await screen.findAllByRole('link', { name: 'Continue' })
    expect(links.map((link) => link.getAttribute('href'))).toEqual([
      '/courses/c1',
      '/courses/c2/lectures/l7',
    ])
  })

  it('explains when the server cannot list the library, and loads it again on retry', async () => {
    let isServerBack = false
    respondWith({
      '/api/courses': () =>
        isServerBack ? Response.json([COURSE]) : new Response(null, { status: 500 }),
    })
    renderPage()
    expect(await screen.findByRole('alert')).toHaveTextContent('Could not load the library')
    isServerBack = true
    fireEvent.click(screen.getByRole('button', { name: 'Reload' }))
    expect(await screen.findByText('Backend Fundamentals')).toBeInTheDocument()
  })

  it('points to the downloads page when the library is empty', async () => {
    respondWith({ '/api/courses': () => Response.json([]) })
    renderPage()
    expect(await screen.findByText('No courses yet')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Download a course' })).toHaveAttribute(
      'href',
      '/downloads',
    )
  })
})

describe('LibraryPage after a failed background refresh', () => {
  it('keeps showing the courses it already has', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    let failNext = false
    vi.spyOn(globalThis, 'fetch').mockImplementation((input) => {
      const path = new URL(input instanceof Request ? input.url : String(input)).pathname
      if (path === '/api/library/tools') return Promise.resolve(Response.json([]))
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
    expect(await screen.findByText(/Could not refresh/)).toBeInTheDocument()
    expect(screen.getByText('Backend Fundamentals')).toBeInTheDocument()
  })
})
