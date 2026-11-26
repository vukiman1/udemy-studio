import type { Job, JobStatus } from '../../api/types'

const ACTIVE_STATUSES: readonly JobStatus[] = ['queued', 'running']

export const JOB_STATUS_LABELS: Readonly<Record<JobStatus, string>> = {
  queued: 'Queued',
  running: 'Downloading',
  succeeded: 'Done',
  failed: 'Failed',
  cancelled: 'Cancelled',
}

const JOB_STATUSES: readonly string[] = Object.keys(JOB_STATUS_LABELS)

export function upsertJob(jobs: readonly Job[], update: Job): Job[] {
  const others = jobs.filter((job) => job.id !== update.id)
  return [...others, update].sort((left, right) => right.createdAt.localeCompare(left.createdAt))
}

export function isActive(job: Job): boolean {
  return ACTIVE_STATUSES.includes(job.status)
}

export function parseJobEvent(data: string): Job | undefined {
  let candidate: unknown
  try {
    candidate = JSON.parse(data)
  } catch {
    return undefined
  }
  if (typeof candidate !== 'object' || candidate === null) return undefined
  const record = candidate as Record<string, unknown>
  const progress = record.progress as Record<string, unknown> | undefined
  const isJob =
    typeof record.id === 'string' &&
    typeof record.courseUrl === 'string' &&
    typeof record.createdAt === 'string' &&
    typeof record.status === 'string' &&
    JOB_STATUSES.includes(record.status) &&
    Array.isArray(record.failures) &&
    typeof progress?.totalLectures === 'number' &&
    typeof progress.finishedCount === 'number'
  return isJob ? (candidate as Job) : undefined
}
