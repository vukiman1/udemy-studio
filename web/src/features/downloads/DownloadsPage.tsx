import type { ReactElement } from 'react'
import { Download } from 'lucide-react'
import { DOWNLOADS_UNAVAILABLE_STATUS, useJobs } from '../../api/download-queries'
import { ApiRequestError } from '../../api/queries'
import { EmptyState } from '../../ui/EmptyState'
import { JobCard } from './JobCard'
import { NewJobForm } from './NewJobForm'
import { PastedLogins } from './PastedLogins'
import { ProfilePicker } from './ProfilePicker'
import { useJobEvents } from './use-job-events'

export function DownloadsPage(): ReactElement {
  const jobs = useJobs()
  const isUnavailable =
    jobs.error instanceof ApiRequestError && jobs.error.status === DOWNLOADS_UNAVAILABLE_STATUS
  useJobEvents(jobs.data !== undefined)

  return (
    <main className="grid max-w-[1056px] content-start gap-5 px-5 py-8 md:px-10 md:py-10">
      <h1 className="text-[30px] font-semibold">Downloads</h1>
      {isUnavailable ? (
        <EmptyState
          icon={Download}
          title="Downloads unavailable on this machine"
          body="The server needs yt-dlp and aria2c. Install them or pass --yt-dlp and --aria2c."
        />
      ) : (
        <>
          <section
            className="grid gap-4 rounded-[20px] bg-card p-6"
            aria-labelledby="udemy-account-title"
          >
            <h2 id="udemy-account-title" className="text-base font-semibold">
              Udemy account
            </h2>
            <ProfilePicker />
            <PastedLogins />
          </section>
          <NewJobForm />
          <section className="grid gap-3" aria-labelledby="job-history-title">
            <h2 id="job-history-title" className="text-[19px] font-semibold">
              Download history
            </h2>
            {jobs.isError && (
              <p className="text-sm text-danger" role="alert">
                Could not load download history ({jobs.error.message}).
              </p>
            )}
            {jobs.data?.length === 0 && <p className="text-sm text-muted">No downloads yet.</p>}
            {jobs.data?.map((job) => (
              <JobCard key={job.id} job={job} />
            ))}
          </section>
        </>
      )}
    </main>
  )
}
