import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { AppShell } from './AppShell'

type Role = 'library' | 'cloud'

function renderShell(role: Role, path: string): void {
  vi.spyOn(globalThis, 'fetch').mockImplementation(() =>
    Promise.resolve(Response.json({ status: 'ok', version: 'test', role })),
  )
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[path]}>
        <Routes>
          <Route element={<AppShell />}>
            <Route path="*" element={<p>page</p>} />
          </Route>
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('AppShell', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('marks only the current page in both the sidebar and the tab bar', async () => {
    renderShell('library', '/downloads')
    await screen.findAllByRole('link', { name: 'Tải khoá học' })
    for (const nav of screen.getAllByRole('navigation', { name: 'Điều hướng chính' })) {
      const current = within(nav)
        .getAllByRole('link')
        .filter((link) => link.getAttribute('aria-current') === 'page')
      expect(current.map((link) => link.textContent)).toEqual(['Tải khoá học'])
    }
  })

  it('leaves downloads and AI settings out of the cloud copy', async () => {
    renderShell('cloud', '/')
    await screen.findAllByRole('link', { name: 'Lưu trữ' })
    expect(screen.queryByRole('link', { name: 'Tải khoá học' })).toBeNull()
    expect(screen.queryByRole('link', { name: 'Cài đặt AI' })).toBeNull()
  })
})
