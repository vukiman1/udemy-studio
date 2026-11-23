import { afterEach, describe, expect, it, vi } from 'vitest'
import { apiClient } from './client'

afterEach(() => {
  vi.restoreAllMocks()
})

describe('apiClient', () => {
  it('sends array query parameters comma-separated, as api/openapi.yaml declares', async () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(Response.json([]))
    await apiClient.GET('/api/lectures/{lectureId}/cues', {
      params: { path: { lectureId: 'l1' }, query: { langs: ['vi', 'en'] } },
    })
    const request = fetchSpy.mock.calls[0]?.[0]
    const url = new URL(request instanceof Request ? request.url : String(request))
    expect(url.searchParams.getAll('langs')).toEqual(['vi,en'])
  })
})
