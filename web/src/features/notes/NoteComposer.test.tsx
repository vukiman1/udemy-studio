import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { NoteComposer } from './NoteComposer'

afterEach(() => {
  vi.restoreAllMocks()
})

describe('NoteComposer', () => {
  it('saves the note at the moment of its first keystroke, not when it is saved', async () => {
    const fetchSpy = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValue(Response.json({ id: 1 }, { status: 201 }))
    const video = document.createElement('video')
    Object.defineProperty(video, 'currentTime', { writable: true, value: 75 })
    const { getByRole } = render(
      <QueryClientProvider client={new QueryClient()}>
        <NoteComposer lectureId="l1" videoRef={{ current: video }} />
      </QueryClientProvider>,
    )
    const textarea = getByRole('textbox')
    fireEvent.change(textarea, { target: { value: 'B' } })
    video.currentTime = 90
    fireEvent.change(textarea, { target: { value: 'Bắt tay ba bước' } })
    fireEvent.click(getByRole('button'))

    await vi.waitFor(async () => {
      const request = fetchSpy.mock.calls[0]?.[0]
      if (!(request instanceof Request)) throw new Error('the note was not sent')
      expect(await request.clone().json()).toEqual({ atSeconds: 75, body: 'Bắt tay ba bước' })
    })
  })
})
