import type { ReactElement } from 'react'
import type { CompressionRun } from '../../api/types'
import { ProgressBar } from '../../app/ProgressBar'
import { RunInBackgroundButton } from '../../app/RunInBackgroundButton'
import { formatBytes } from './format'

interface CompressionProgressProps {
  readonly run: CompressionRun
  readonly isCancelling: boolean
  readonly onCancel: () => void
  readonly onRunInBackground: () => void
}

const PERCENT = 100

export function CompressionProgress({
  run,
  isCancelling,
  onCancel,
  onRunInBackground,
}: CompressionProgressProps): ReactElement {
  const { current } = run
  const currentPercent =
    current === undefined || current.lengthSeconds === 0
      ? 0
      : Math.min(PERCENT, Math.round((current.encodedSeconds / current.lengthSeconds) * PERCENT))

  return (
    <div className="grid gap-2">
      <p className="m-0 text-[0.9rem]">
        {run.state === 'queued'
          ? 'Đang chờ khoá khác nén xong…'
          : `Đang nén ${String(Math.min(run.finished + 1, run.total))}/${String(run.total)}: ${current?.lectureTitle ?? '…'} (${String(currentPercent)}%)`}
      </p>
      <ProgressBar
        completed={run.finished * PERCENT + currentPercent}
        total={run.total * PERCENT}
      />
      {run.bytesBefore > 0 && (
        <p className="m-0 text-[0.85rem] text-muted">
          Đã giảm {formatBytes(run.bytesBefore)} → {formatBytes(run.bytesAfter)}
        </p>
      )}
      <div className="flex gap-3">
        <RunInBackgroundButton onClick={onRunInBackground} />
        <button type="button" disabled={isCancelling} onClick={onCancel}>
          Dừng nén
        </button>
      </div>
    </div>
  )
}
