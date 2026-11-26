import { useState } from 'react'
import type { StorageDraft } from '../../api/storage-queries'
import type { StorageSettings } from '../../api/types'

export interface StorageDraftEditor {
  readonly draft: StorageDraft
  readonly saved: StorageSettings
  readonly isEditingKeys: boolean
  readonly canKeepSavedKeys: boolean
  readonly isComplete: boolean
  readonly isDirty: boolean
  readonly typeUrl: (url: string) => void
  readonly typeAccessKeyId: (accessKeyId: string) => void
  readonly typeSecretAccessKey: (secretAccessKey: string) => void
  readonly startEditingKeys: () => void
  readonly keepSavedKeys: () => void
}

export function useStorageDraft(saved: StorageSettings): StorageDraftEditor {
  const savedUrl = saved.url ?? ''
  const [draft, setDraft] = useState<StorageDraft>({
    url: savedUrl,
    accessKeyId: '',
    secretAccessKey: '',
  })
  const [hasAskedForNewKeys, setHasAskedForNewKeys] = useState(false)
  const isEditingKeys = !saved.configured || hasAskedForNewKeys
  const hasBothKeys = draft.accessKeyId.trim() !== '' && draft.secretAccessKey.trim() !== ''

  return {
    draft,
    saved,
    isEditingKeys,
    canKeepSavedKeys: saved.configured && hasAskedForNewKeys,
    isComplete: draft.url.trim() !== '' && (!isEditingKeys || hasBothKeys),
    isDirty: isEditingKeys || draft.url.trim() !== savedUrl,
    typeUrl: (url) => {
      setDraft((current) => ({ ...current, url }))
    },
    typeAccessKeyId: (accessKeyId) => {
      setDraft((current) => ({ ...current, accessKeyId }))
    },
    typeSecretAccessKey: (secretAccessKey) => {
      setDraft((current) => ({ ...current, secretAccessKey }))
    },
    startEditingKeys: () => {
      setHasAskedForNewKeys(true)
    },
    keepSavedKeys: () => {
      setHasAskedForNewKeys(false)
      setDraft((current) => ({ ...current, accessKeyId: '', secretAccessKey: '' }))
    },
  }
}
