import { useState } from 'react'
import type { AIDraft } from '../../api/ai-queries'
import type { AIProvider, AIProviderId, AISavedProvider, AISettings } from '../../api/types'

export interface AIDraftEditor {
  readonly draft: AIDraft
  readonly provider: AIProvider
  // The draft provider's saved key, if it has one.
  readonly saved: AISavedProvider | undefined
  readonly isInUse: boolean
  readonly isEditingKey: boolean
  // True while replacing the saved key, which can then be kept after all.
  readonly canKeepSavedKey: boolean
  readonly isDirty: boolean
  readonly chooseProvider: (provider: AIProviderId) => void
  readonly typeKey: (apiKey: string) => void
  readonly startEditingKey: () => void
  readonly keepSavedKey: () => void
}

// The form edits a draft of one provider's settings; keys are only ever typed, never read back.
export function useAIDraft(
  providers: readonly AIProvider[],
  defaultProvider: AIProvider,
  settings: AISettings,
): AIDraftEditor {
  const savedOf = (id: AIProviderId): AISavedProvider | undefined =>
    settings.providers.find((saved) => saved.provider === id)
  const draftFor = (provider: AIProvider): AIDraft => ({ provider: provider.id, apiKey: '' })
  const startProvider =
    providers.find((candidate) => candidate.id === settings.provider) ?? defaultProvider
  const [draft, setDraft] = useState<AIDraft>(draftFor(startProvider))
  const [hasAskedForNewKey, setHasAskedForNewKey] = useState(false)
  const provider = providers.find((candidate) => candidate.id === draft.provider) ?? defaultProvider
  const saved = savedOf(draft.provider)
  const isInUse = settings.configured && settings.provider === draft.provider

  return {
    draft,
    provider,
    saved,
    isInUse,
    isEditingKey: saved === undefined || hasAskedForNewKey,
    canKeepSavedKey: saved !== undefined && hasAskedForNewKey,
    isDirty: !isInUse || saved === undefined || draft.apiKey.trim() !== '',
    chooseProvider: (next) => {
      const nextProvider = providers.find((candidate) => candidate.id === next) ?? provider
      setDraft(draftFor(nextProvider))
      setHasAskedForNewKey(false)
    },
    typeKey: (apiKey) => {
      setDraft((current) => ({ ...current, apiKey }))
    },
    startEditingKey: () => {
      setHasAskedForNewKey(true)
    },
    keepSavedKey: () => {
      setHasAskedForNewKey(false)
      setDraft((current) => ({ ...current, apiKey: '' }))
    },
  }
}
