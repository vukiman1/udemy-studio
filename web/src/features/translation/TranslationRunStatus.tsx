import type { ReactElement } from 'react'
import type { TranslationRun } from '../../api/types'
import { ProgressBar } from '../../app/ProgressBar'
import { formatUsd } from './format-usd'

interface TranslationRunStatusProps {
  readonly run: TranslationRun
}

const STATE_TEXT: Readonly<Record<TranslationRun['state'], string>> = {
  queued: 'Đang chờ khoá khác dịch xong…',
  running: 'Đang dịch',
  done: 'Dịch xong.',
  cancelled: 'Đã dừng.',
  failed: 'Đã dừng vì lỗi.',
}

export function TranslationRunStatus({ run }: TranslationRunStatusProps): ReactElement {
  const translated = run.finished - run.failed.length
  return (
    <div className="grid gap-1 text-[0.9rem]">
      <p className="m-0">
        {run.state === 'running'
          ? `${STATE_TEXT.running} ${String(Math.min(run.finished + 1, run.total))}/${String(run.total)}: ${run.currentLecture ?? '…'}`
          : `${STATE_TEXT[run.state]} ${String(translated)}/${String(run.total)} bài có phụ đề mới.`}{' '}
        <span className="text-muted">Đã tốn {formatUsd(run.costUsd)}.</span>
      </p>
      {run.state === 'running' && <ProgressBar completed={run.finished} total={run.total} />}
      {run.error !== undefined && (
        <p className="m-0 text-danger" role="alert">
          {run.error}
        </p>
      )}
      {run.failed.map((outcome) => (
        <p key={outcome.path} className="m-0 text-danger">
          Không dịch được: {outcome.lectureTitle} ({outcome.reason})
        </p>
      ))}
    </div>
  )
}
