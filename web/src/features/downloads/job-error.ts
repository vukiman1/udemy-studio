import type { Job } from '../../api/types'

type JobErrorKind = NonNullable<Job['errorKind']>

const ERROR_ADVICE: Readonly<Record<Exclude<JobErrorKind, 'other'>, string>> = {
  login: 'Udemy login expired. Paste a new token and download again.',
  course_not_found:
    'Course not found in this Udemy account. Check the link, or open the course once in a browser and download again.',
  blocked:
    'Udemy is blocking the app (Cloudflare). Try later; if it persists, update udemy-studio.',
  chrome_access:
    'Could not read the Chrome login. Grant access or paste a token, then download again.',
  interrupted: 'The server stopped mid-way. Download again to finish.',
}

export function jobErrorText(job: Job): string | undefined {
  if (job.error === undefined) return undefined
  if (job.errorKind === undefined || job.errorKind === 'other') return job.error
  return ERROR_ADVICE[job.errorKind]
}
