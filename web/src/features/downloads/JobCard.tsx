import { useMutation } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { cancelJob } from '../../api/download-queries'
import type { Job } from '../../api/types'
import { ProgressBar } from '../../app/ProgressBar'
import { JOB_STATUS_LABELS, isActive } from './job-list'

const TIME_FORMAT = new Intl.DateTimeFormat('vi-VN', { dateStyle: 'short', timeStyle: 'short' })

const STATUS_STYLES: Readonly<Record<Job['status'], string>> = {
  queued: 'text-muted',
  running: 'text-accent',
  succeeded: 'text-accent',
  failed: 'text-danger',
  cancelled: 'text-muted',
}

interface JobCardProps {
  readonly job: Job
}

export function JobCard({ job }: JobCardProps): ReactElement {
  const cancel = useMutation({ mutationFn: cancelJob })
  const { progress } = job

  return (
    <article className="grid gap-2 rounded-xl border border-line p-4">
      <header className="flex flex-wrap items-baseline justify-between gap-2">
        <h3 className="font-semibold wrap-anywhere">{progress.courseTitle || job.courseUrl}</h3>
        <span className={`text-[0.9rem] font-semibold ${STATUS_STYLES[job.status]}`}>
          {JOB_STATUS_LABELS[job.status]}
        </span>
      </header>
      <p className="text-[0.85rem] text-muted">
        {TIME_FORMAT.format(new Date(job.createdAt))}
        {job.chapters === undefined ? '' : ` · chapter ${job.chapters}`}
        {progress.totalLectures > 0 &&
          ` · ${String(progress.finishedCount)}/${String(progress.totalLectures)} bài`}
      </p>
      {progress.totalLectures > 0 && (
        <ProgressBar completed={progress.finishedCount} total={progress.totalLectures} />
      )}
      {progress.currentLecture !== undefined && (
        <p className="text-[0.9rem] wrap-anywhere">Đang xử lý: {progress.currentLecture}</p>
      )}
      {job.error !== undefined && (
        <p className="text-[0.9rem] wrap-anywhere text-danger">{job.error}</p>
      )}
      {job.failures.length > 0 && (
        <details className="text-[0.9rem]">
          <summary className="cursor-pointer text-danger">
            {job.failures.length} mục lỗi, tải lại để thử lại
          </summary>
          <ul className="mt-1 list-disc pl-5 wrap-anywhere">
            {job.failures.map((failure) => (
              <li key={failure}>{failure}</li>
            ))}
          </ul>
        </details>
      )}
      {isActive(job) && (
        <div className="flex items-center gap-3">
          <button
            type="button"
            disabled={cancel.isPending}
            onClick={() => {
              cancel.mutate(job.id)
            }}
          >
            Huỷ
          </button>
          {cancel.isError && <p className="text-danger">{cancel.error.message}</p>}
        </div>
      )}
    </article>
  )
}
