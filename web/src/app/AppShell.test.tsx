import { render, screen, within } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { describe, expect, it } from 'vitest'
import { AppShell } from './AppShell'

describe('AppShell', () => {
  it('marks only the current page in both the sidebar and the tab bar', () => {
    render(
      <MemoryRouter initialEntries={['/downloads']}>
        <Routes>
          <Route element={<AppShell />}>
            <Route path="*" element={<p>page</p>} />
          </Route>
        </Routes>
      </MemoryRouter>,
    )
    for (const nav of screen.getAllByRole('navigation', { name: 'Điều hướng chính' })) {
      const current = within(nav)
        .getAllByRole('link')
        .filter((link) => link.getAttribute('aria-current') === 'page')
      expect(current.map((link) => link.textContent)).toEqual(['Tải khoá học'])
    }
  })
})
