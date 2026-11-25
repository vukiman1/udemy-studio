import { useMutation } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { cancelJob } from '../../api/download-queries'
import type { Job } from '../../api/types'
import { ProgressBar } from '../../app/ProgressBar'
import { Button, ButtonSize, ButtonVariant } from '../../ui/Button'
import { JOB_STATUS_LABELS, isActive } from './job-list'

const TIME_FORMAT = new Intl.DateTimeFormat('vi-VN', { dateStyle: 'short', timeStyle: 'short' })

const STATUS_PILLS: Readonly<Record<Job['status'], string>> = {
  queued: 'bg-button text-muted',
  running: 'bg-accent/20 text-accent-soft',
  succeeded: 'bg-success-soft text-success',
  failed: 'bg-danger/15 text-danger',
  cancelled: 'bg-button text-muted',
}

interface JobCardProps {
  readonly job: Job
}

export function JobCard({ job }: JobCardProps): ReactElement {
  const cancel = useMutation({ mutationFn: cancelJob })
  const { progress } = job

  return (
    <article className="grid gap-2.5 rounded-2xl bg-card px-5 py-[18px]">
      <header className="flex items-center gap-3">
        <h3 className="flex-1 text-[15px] font-semibold wrap-anywhere">
          {progress.courseTitle || job.courseUrl}
        </h3>
        <span
          className={`shrink-0 rounded-full px-2.5 py-1 text-xs font-semibold ${STATUS_PILLS[job.status]}`}
        >
          {JOB_STATUS_LABELS[job.status]}
        </span>
        {isActive(job) && (
          <Button
            variant={ButtonVariant.Ghost}
            size={ButtonSize.Small}
            disabled={cancel.isPending}
            onClick={() => {
              cancel.mutate(job.id)
            }}
          >
            Huỷ
          </Button>
        )}
      </header>
      <p className="text-xs text-muted">
        {TIME_FORMAT.format(new Date(job.createdAt))}
        {job.chapters === undefined ? '' : ` · chapter ${job.chapters}`}
        {progress.totalLectures > 0 &&
          ` · ${String(progress.finishedCount)}/${String(progress.totalLectures)} bài`}
      </p>
      {isActive(job) && progress.totalLectures > 0 && (
        <ProgressBar
          className="h-2 rounded"
          completed={progress.finishedCount}
          total={progress.totalLectures}
        />
      )}
      {progress.currentLecture !== undefined && (
        <p className="text-[13px] wrap-anywhere">Đang xử lý: {progress.currentLecture}</p>
      )}
      {job.error !== undefined && (
        <p className="text-[13px] wrap-anywhere text-danger">{job.error}</p>
      )}
      {cancel.isError && (
        <p className="text-[13px] text-danger" role="alert">
          Không huỷ được: {cancel.error.message}
        </p>
      )}
      {job.failures.length > 0 && (
        <details className="group text-[13px]">
          <summary className="cursor-pointer list-none font-medium text-danger [&::-webkit-details-marker]:hidden">
            <span className="inline-block transition-transform group-open:rotate-90">▸</span>{' '}
            {job.failures.length} mục lỗi, tải lại để thử lại
          </summary>
          <ul className="mt-1.5 grid list-disc gap-1 pl-5 wrap-anywhere">
            {job.failures.map((failure) => (
              <li key={failure}>{failure}</li>
            ))}
          </ul>
        </details>
      )}
    </article>
  )
}
