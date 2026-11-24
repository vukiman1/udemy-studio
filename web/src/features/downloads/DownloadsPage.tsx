import { ChevronIcon } from '@videojs/react/icons'
import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { DOWNLOADS_UNAVAILABLE_STATUS, useJobs } from '../../api/download-queries'
import { ApiRequestError } from '../../api/queries'
import { PageStatus } from '../../app/PageStatus'
import { JobCard } from './JobCard'
import { NewJobForm } from './NewJobForm'
import { PastedLogins } from './PastedLogins'
import { ProfilePicker } from './ProfilePicker'
import { useJobEvents } from './use-job-events'

export function DownloadsPage(): ReactElement {
  const jobs = useJobs()
  const unavailable =
    jobs.error instanceof ApiRequestError && jobs.error.status === DOWNLOADS_UNAVAILABLE_STATUS
      ? jobs.error
      : undefined
  useJobEvents(jobs.data !== undefined)

  return (
    <main className="mx-auto grid max-w-[900px] gap-4 px-4 py-6">
      <header className="flex items-center gap-3">
        <Link to="/" className="inline-flex items-center gap-1 text-muted">
          <ChevronIcon className="size-4.5 -scale-x-100" />
          Thư viện
        </Link>
        <h1 className="text-2xl font-bold">Tải khoá học</h1>
      </header>
      {unavailable !== undefined ? (
        <PageStatus tone="error" message={unavailable.message} />
      ) : (
        <>
          <section className="grid gap-4 rounded-xl border border-line bg-surface p-4">
            <h2 className="font-semibold">Tài khoản Udemy</h2>
            <ProfilePicker />
            <PastedLogins />
          </section>
          <NewJobForm />
          <section className="grid gap-3">
            <h2 className="font-semibold">Lịch sử tải</h2>
            {jobs.isError && <PageStatus tone="error" message={jobs.error.message} />}
            {jobs.data?.length === 0 && <p className="text-muted">Chưa có lần tải nào.</p>}
            {jobs.data?.map((job) => (
              <JobCard key={job.id} job={job} />
            ))}
          </section>
        </>
      )}
    </main>
  )
}
