import { QueryClient } from '@tanstack/react-query'

const STALE_AFTER_MS = 30_000

export function createQueryClient(): QueryClient {
  return new QueryClient({ defaultOptions: { queries: { staleTime: STALE_AFTER_MS, retry: 1 } } })
}
