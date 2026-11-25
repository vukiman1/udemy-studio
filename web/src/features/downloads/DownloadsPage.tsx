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
      <h1 className="text-[30px] font-semibold">Tải khoá học</h1>
      {isUnavailable ? (
        <EmptyState
          icon={Download}
          title="Chưa tải được khoá học trên máy này"
          body="Máy chạy server cần yt-dlp và aria2c. Cài chúng, hoặc chỉ đường dẫn bằng --yt-dlp và --aria2c."
        />
      ) : (
        <>
          <section
            className="grid gap-4 rounded-[20px] bg-card p-6"
            aria-labelledby="udemy-account-title"
          >
            <h2 id="udemy-account-title" className="text-base font-semibold">
              Tài khoản Udemy
            </h2>
            <ProfilePicker />
            <PastedLogins />
          </section>
          <NewJobForm />
          <section className="grid gap-3" aria-labelledby="job-history-title">
            <h2 id="job-history-title" className="text-[19px] font-semibold">
              Lịch sử tải
            </h2>
            {jobs.isError && (
              <p className="text-sm text-danger" role="alert">
                Không tải được lịch sử tải ({jobs.error.message}).
              </p>
            )}
            {jobs.data?.length === 0 && <p className="text-sm text-muted">Chưa có lần tải nào.</p>}
            {jobs.data?.map((job) => (
              <JobCard key={job.id} job={job} />
            ))}
          </section>
        </>
      )}
    </main>
  )
}
