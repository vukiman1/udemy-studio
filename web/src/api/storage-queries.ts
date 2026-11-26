import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { apiClient } from './client'
import { ApiRequestError, QUERY_KEYS } from './queries'
import type { StorageCheck, StorageSettings } from './types'

export interface StorageDraft {
  readonly url: string
  readonly accessKeyId: string
  readonly secretAccessKey: string
}

interface StorageDraftBody {
  readonly url: string
  readonly accessKeyId?: string
  readonly secretAccessKey?: string
}

function draftBody(draft: StorageDraft): StorageDraftBody {
  const hasTypedKeys = draft.accessKeyId.trim() !== '' || draft.secretAccessKey.trim() !== ''
  return hasTypedKeys
    ? { url: draft.url, accessKeyId: draft.accessKeyId, secretAccessKey: draft.secretAccessKey }
    : { url: draft.url }
}

export function useStorageSettings(): UseQueryResult<StorageSettings> {
  return useQuery({
    queryKey: QUERY_KEYS.storageSettings,
    queryFn: async () => {
      const { data, response } = await apiClient.GET('/api/storage')
      if (data === undefined) throw new ApiRequestError('load storage settings', response.status)
      return data
    },
  })
}

export async function saveStorageSettings(draft: StorageDraft): Promise<StorageSettings> {
  const { data, error, response } = await apiClient.PUT('/api/storage', { body: draftBody(draft) })
  if (data === undefined)
    throw new ApiRequestError('save storage settings', response.status, error.detail)
  return data
}

export async function forgetStorageSettings(): Promise<void> {
  const { response } = await apiClient.DELETE('/api/storage')
  if (!response.ok) throw new ApiRequestError('forget storage settings', response.status)
}

export async function checkStorage(draft: StorageDraft): Promise<StorageCheck> {
  const { data, error, response } = await apiClient.POST('/api/storage/check', {
    body: draftBody(draft),
  })
  if (data === undefined)
    throw new ApiRequestError('check the storage connection', response.status, error.detail)
  return data
}
