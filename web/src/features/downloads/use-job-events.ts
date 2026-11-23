import { useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { QUERY_KEYS } from '../../api/queries'
import type { Job } from '../../api/types'
import { parseJobEvent, upsertJob } from './job-list'

const EVENTS_URL = '/api/events'
const JOB_EVENT = 'job'

export function useJobEvents(isEnabled: boolean): void {
  const queryClient = useQueryClient()

  useEffect(() => {
    if (!isEnabled) return
    const source = new EventSource(EVENTS_URL)
    const onJob = (event: MessageEvent<string>): void => {
      const job = parseJobEvent(event.data)
      if (job === undefined) return
      queryClient.setQueryData<Job[]>(QUERY_KEYS.jobs, (current) => upsertJob(current ?? [], job))
      if (job.status === 'succeeded') {
        void queryClient.invalidateQueries({ queryKey: QUERY_KEYS.courses })
      }
    }
    source.addEventListener(JOB_EVENT, onJob)
    return () => {
      source.removeEventListener(JOB_EVENT, onJob)
      source.close()
    }
  }, [isEnabled, queryClient])
}
