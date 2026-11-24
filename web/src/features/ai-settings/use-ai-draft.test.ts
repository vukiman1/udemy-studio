import { act, renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { AIProvider, AISettings } from '../../api/types'
import { useAIDraft } from './use-ai-draft'

const CLAUDE: AIProvider = {
  id: 'claude',
  label: 'Claude (Anthropic)',
  model: { id: 'claude-haiku-5-5', label: 'Claude Haiku 5.5' },
  inputUsdPerMillion: 0.1,
  outputUsdPerMillion: 0.5,
}

const SAVED: AISettings = {
  configured: true,
  provider: 'claude',
  enabled: true,
  providers: [{ provider: 'claude', keyHint: 'sk-a••••••••GAAA' }],
}

describe('useAIDraft', () => {
  it('goes back to the saved key when changing it is cancelled', () => {
    const { result } = renderHook(() => useAIDraft([CLAUDE], CLAUDE, SAVED))
    act(() => {
      result.current.startEditingKey()
    })
    act(() => {
      result.current.typeKey('sk-half-typed')
    })
    expect(result.current.isDirty).toBe(true)

    act(() => {
      result.current.keepSavedKey()
    })
    expect(result.current.isEditingKey).toBe(false)
    expect(result.current.draft.apiKey).toBe('')
    expect(result.current.isDirty).toBe(false)
  })
})
