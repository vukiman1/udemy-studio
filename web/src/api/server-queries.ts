import { useQuery } from '@tanstack/react-query'
import { apiClient } from './client'
import { ApiRequestError, QUERY_KEYS } from './queries'

type ServerRole = 'library' | 'cloud'

function useServerRole(): ServerRole | undefined {
  const health = useQuery({
    queryKey: QUERY_KEYS.server,
    staleTime: Infinity,
    queryFn: async () => {
      const { data, response } = await apiClient.GET('/healthz')
      if (data === undefined) throw new ApiRequestError('read the server role', response.status)
      return data
    },
  })
  return health.data?.role
}

export function useIsLibraryCopy(): boolean {
  return useServerRole() === 'library'
}

export function useIsCloudCopy(): boolean {
  return useServerRole() === 'cloud'
}

export async function refreshLibrary(): Promise<void> {
  const { error, response } = await apiClient.POST('/api/library/refresh')
  if (!response.ok) throw new ApiRequestError('refresh the library', response.status, error?.detail)
}
