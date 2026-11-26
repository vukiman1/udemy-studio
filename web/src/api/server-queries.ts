import { useQuery } from '@tanstack/react-query'
import { apiClient } from './client'
import { ApiRequestError, QUERY_KEYS } from './queries'

export function useIsLibraryCopy(): boolean {
  const health = useQuery({
    queryKey: QUERY_KEYS.server,
    staleTime: Infinity,
    queryFn: async () => {
      const { data, response } = await apiClient.GET('/healthz')
      if (data === undefined) throw new ApiRequestError('read the server role', response.status)
      return data
    },
  })
  return health.data?.role === 'library'
}
