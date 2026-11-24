import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { apiClient } from './client'
import { ApiRequestError, QUERY_KEYS } from './queries'
import type { ChromeProfile, Job, JobRequest, LoginStatus, PastedLogin } from './types'

export const DOWNLOADS_UNAVAILABLE_STATUS = 503

export function useChromeProfiles(): UseQueryResult<ChromeProfile[]> {
  return useQuery({
    queryKey: QUERY_KEYS.chromeProfiles,
    queryFn: async () => {
      const { data, error, response } = await apiClient.GET('/api/downloads/profiles')
      if (data === undefined)
        throw new ApiRequestError('list Chrome profiles', response.status, error.detail)
      return data
    },
  })
}

export function useLoginStatus(): UseQueryResult<LoginStatus> {
  return useQuery({
    queryKey: QUERY_KEYS.loginStatus,
    queryFn: async () => {
      const { data, error, response } = await apiClient.GET('/api/downloads/status')
      if (data === undefined)
        throw new ApiRequestError('load login status', response.status, error.detail)
      return data
    },
  })
}

// Job changes arrive over SSE (with a full snapshot on every reconnect), so a refetch could only overwrite newer events.
export function useJobs(): UseQueryResult<Job[]> {
  return useQuery({
    queryKey: QUERY_KEYS.jobs,
    staleTime: Infinity,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    queryFn: async () => {
      const { data, error, response } = await apiClient.GET('/api/jobs')
      if (data === undefined) throw new ApiRequestError('list jobs', response.status, error.detail)
      return data
    },
  })
}

export function usePastedLogins(): UseQueryResult<PastedLogin[]> {
  return useQuery({
    queryKey: QUERY_KEYS.pastedLogins,
    queryFn: async () => {
      const { data, error, response } = await apiClient.GET('/api/downloads/tokens')
      if (data === undefined)
        throw new ApiRequestError('list pasted Udemy tokens', response.status, error.detail)
      return data
    },
  })
}

// The server checks the token with the portal before keeping it, so a rejected paste fails here.
export async function savePastedLogin(host: string, token: string): Promise<PastedLogin> {
  const { data, error, response } = await apiClient.PUT('/api/downloads/tokens', {
    body: { host, token },
  })
  if (data === undefined)
    throw new ApiRequestError('save the Udemy token', response.status, error.detail)
  return data
}

export async function forgetPastedLogin(host: string): Promise<void> {
  const { error, response } = await apiClient.DELETE('/api/downloads/tokens/{host}', {
    params: { path: { host } },
  })
  if (!response.ok)
    throw new ApiRequestError(`forget the token for ${host}`, response.status, error?.detail)
}

export async function selectChromeProfile(dir: string): Promise<void> {
  const { error, response } = await apiClient.PUT('/api/downloads/profile', { body: { dir } })
  if (!response.ok)
    throw new ApiRequestError(`select profile ${dir}`, response.status, error?.detail)
}

export async function createJob(request: JobRequest): Promise<Job> {
  const { data, error, response } = await apiClient.POST('/api/jobs', { body: request })
  if (data === undefined) throw new ApiRequestError('queue download', response.status, error.detail)
  return data
}

export async function cancelJob(jobId: string): Promise<void> {
  const { error, response } = await apiClient.DELETE('/api/jobs/{jobId}', {
    params: { path: { jobId } },
  })
  if (!response.ok) throw new ApiRequestError(`cancel job ${jobId}`, response.status, error?.detail)
}
