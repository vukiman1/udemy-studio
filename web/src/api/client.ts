import createClient from 'openapi-fetch'
import type { paths } from './schema'

function fetchAtCallTime(request: Request): Promise<Response> {
  return globalThis.fetch(request)
}

export const apiClient = createClient<paths>({
  baseUrl: window.location.origin,
  fetch: fetchAtCallTime,
  querySerializer: { array: { style: 'form', explode: false } },
})
