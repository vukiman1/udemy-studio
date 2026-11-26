import type { ReactElement } from 'react'
import type { PublishRun } from '../../api/types'
import { ProgressBar, completionPercent } from '../../app/ProgressBar'
import { formatBytes } from '../compression/format'

interface PublishProgressProps {
  readonly run: PublishRun
}

export function PublishProgress({ run }: PublishProgressProps): ReactElement {
  const handledFiles = run.sentFiles + run.skippedFiles
  return (
    <div className="grid gap-2 text-[13px]">
      <p className="flex items-baseline justify-between gap-3">
        <span className="text-[15px] font-semibold">
          Đang đưa lên {handledFiles}/{run.totalFiles} file
        </span>
        <span className="text-muted tabular-nums">
          {completionPercent(run.sentBytes, run.totalBytes)}%
        </span>
      </p>
      <p className="wrap-anywhere text-muted">File hiện tại: {run.currentFile ?? '…'}</p>
      <ProgressBar className="h-2 rounded" completed={run.sentBytes} total={run.totalBytes} />
      <p className="text-xs text-muted">
        {formatBytes(run.sentBytes)} / {formatBytes(run.totalBytes)}
      </p>
    </div>
  )
}
