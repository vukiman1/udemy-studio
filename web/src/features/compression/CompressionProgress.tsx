import type { ReactElement } from 'react'
import type { CompressionRun } from '../../api/types'
import { ProgressBar, completionPercent } from '../../app/ProgressBar'
import { formatBytes } from './format'

interface CompressionProgressProps {
  readonly run: CompressionRun
}

const PERCENT = 100

export function CompressionProgress({ run }: CompressionProgressProps): ReactElement {
  if (run.state === 'queued')
    return <p className="text-[13px] text-muted">Đang chờ khoá khác nén xong…</p>
  const { current } = run
  const currentPercent =
    current === undefined || current.lengthSeconds === 0
      ? 0
      : Math.min(PERCENT, Math.round((current.encodedSeconds / current.lengthSeconds) * PERCENT))
  const overallDone = run.finished * PERCENT + currentPercent
  const overallTotal = run.total * PERCENT

  return (
    <div className="grid gap-2 text-[13px]">
      <p className="flex items-baseline justify-between gap-3">
        <span className="text-[15px] font-semibold">
          Đang nén {Math.min(run.finished + 1, run.total)}/{run.total}
        </span>
        <span className="text-muted tabular-nums">
          {completionPercent(overallDone, overallTotal)}%
        </span>
      </p>
      <p className="wrap-anywhere text-muted">
        Video hiện tại: {current?.lectureTitle ?? '…'} ({currentPercent}%)
      </p>
      <ProgressBar className="h-2 rounded" completed={overallDone} total={overallTotal} />
      {run.bytesBefore > 0 && (
        <p className="text-xs text-muted">
          Đã giảm {formatBytes(run.bytesBefore)} → {formatBytes(run.bytesAfter)}
        </p>
      )}
    </div>
  )
}
