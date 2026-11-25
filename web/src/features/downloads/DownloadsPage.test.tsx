import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { DownloadsPage } from './DownloadsPage'

afterEach(() => {
  vi.restoreAllMocks()
})

describe('DownloadsPage', () => {
  it('tells how to enable downloads when the server lacks yt-dlp or aria2c', async () => {
    vi.spyOn(globalThis, 'fetch').mockImplementation(() =>
      Promise.resolve(Response.json({ detail: 'yt-dlp not found in PATH' }, { status: 503 })),
    )
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={client}>
        <DownloadsPage />
      </QueryClientProvider>,
    )
    expect(await screen.findByRole('status')).toHaveTextContent(
      'Chưa tải được khoá học trên máy này',
    )
    expect(screen.queryByRole('heading', { name: 'Tài khoản Udemy' })).not.toBeInTheDocument()
  })
})
