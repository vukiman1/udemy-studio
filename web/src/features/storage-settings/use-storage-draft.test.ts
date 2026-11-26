import { act, renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { StorageSettings } from '../../api/types'
import { useStorageDraft } from './use-storage-draft'

const SAVED_URL = 'https://abc123.r2.cloudflarestorage.com/studio-dev'

const SAVED: StorageSettings = {
  configured: true,
  url: SAVED_URL,
  accessKeyIdHint: '0123••••••••cdef',
}

describe('useStorageDraft', () => {
  it('lets a new bucket URL be saved with the saved keys', () => {
    const { result } = renderHook(() => useStorageDraft(SAVED))
    expect(result.current.isDirty).toBe(false)

    act(() => {
      result.current.typeUrl('https://abc123.r2.cloudflarestorage.com/studio')
    })
    expect(result.current.isDirty).toBe(true)
    expect(result.current.isComplete).toBe(true)
  })

  it('needs both keys before new keys can be saved', () => {
    const { result } = renderHook(() => useStorageDraft(SAVED))
    act(() => {
      result.current.startEditingKeys()
    })
    act(() => {
      result.current.typeAccessKeyId('new-key-id')
    })
    expect(result.current.isComplete).toBe(false)

    act(() => {
      result.current.typeSecretAccessKey('new-secret')
    })
    expect(result.current.isComplete).toBe(true)
  })
})
