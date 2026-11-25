import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { HistoryEntry } from '../../api/types'
import { HistoryPage } from './HistoryPage'

const LEFT_MIDWAY: HistoryEntry = {
  courseId: 'c1',
  courseTitle: 'Backend Fundamentals',
  lectureId: 'l3',
  lectureTitle: 'Course Outline',
  positionSeconds: 765.4,
  completed: false,
  updatedAt: '2026-11-25T08:00:00Z',
}

function renderWithHistory(entries: readonly HistoryEntry[]): void {
  vi.spyOn(globalThis, 'fetch').mockImplementation(() => Promise.resolve(Response.json(entries)))
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <HistoryPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

afterEach(() => {
  vi.restoreAllMocks()
})

describe('HistoryPage', () => {
  it('resumes an unfinished lecture at the moment it was left', async () => {
    renderWithHistory([LEFT_MIDWAY])
    expect(await screen.findByText('Đang học dở · dừng ở 12:45')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Học tiếp' })).toHaveAttribute(
      'href',
      '/courses/c1/lectures/l3?t=765',
    )
  })

  it('reopens a finished lecture from its start', async () => {
    renderWithHistory([{ ...LEFT_MIDWAY, completed: true }])
    expect(await screen.findByText('Đã học xong')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Học tiếp' })).toHaveAttribute(
      'href',
      '/courses/c1/lectures/l3',
    )
  })

  it('points to the library before anything was watched', async () => {
    renderWithHistory([])
    expect(await screen.findByText('Chưa học bài nào')).toBeInTheDocument()
  })
})
