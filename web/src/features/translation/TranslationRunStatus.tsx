import type { ReactElement } from 'react'
import type { TranslationRun } from '../../api/types'
import { ProgressBar, completionPercent } from '../../app/ProgressBar'
import { formatUsd } from './format-usd'

interface TranslationRunStatusProps {
  readonly run: TranslationRun
}

const FINISHED_TEXT: Readonly<
  Record<Exclude<TranslationRun['state'], 'running' | 'queued'>, string>
> = {
  done: 'Dịch xong.',
  cancelled: 'Đã dừng.',
  failed: 'Đã dừng vì lỗi.',
}

export function TranslationRunStatus({ run }: TranslationRunStatusProps): ReactElement {
  return (
    <div className="grid gap-2 text-[13px]">
      <RunSummary run={run} />
      {run.error !== undefined && (
        <p className="text-danger" role="alert">
          {run.error}
        </p>
      )}
      {run.failed.map((outcome) => (
        <p key={outcome.path} className="text-danger">
          Không dịch được: {outcome.lectureTitle} ({outcome.reason})
        </p>
      ))}
    </div>
  )
}

function RunSummary({ run }: TranslationRunStatusProps): ReactElement {
  if (run.state === 'queued') return <p className="text-muted">Đang chờ khoá khác dịch xong…</p>
  if (run.state === 'running')
    return (
      <>
        <p className="flex items-baseline justify-between gap-3">
          <span className="text-[15px] font-semibold">
            Đang dịch {Math.min(run.finished + 1, run.total)}/{run.total}
          </span>
          <span className="text-muted tabular-nums">
            {completionPercent(run.finished, run.total)}%
          </span>
        </p>
        <p className="wrap-anywhere text-muted">Bài hiện tại: {run.currentLecture ?? '…'}</p>
        <ProgressBar className="h-2 rounded" completed={run.finished} total={run.total} />
        <p className="text-xs text-muted">Đã tốn {formatUsd(run.costUsd)}</p>
      </>
    )
  const translated = run.finished - run.failed.length
  return (
    <p>
      {FINISHED_TEXT[run.state]} {translated}/{run.total} bài có phụ đề mới.{' '}
      <span className="text-muted">Đã tốn {formatUsd(run.costUsd)}.</span>
    </p>
  )
}
