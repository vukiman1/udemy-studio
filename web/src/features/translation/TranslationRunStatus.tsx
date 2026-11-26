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
  done: 'Done.',
  cancelled: 'Stopped.',
  failed: 'Stopped on an error.',
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
          Not translated: {outcome.lectureTitle} ({outcome.reason})
        </p>
      ))}
    </div>
  )
}

function RunSummary({ run }: TranslationRunStatusProps): ReactElement {
  if (run.state === 'queued')
    return <p className="text-muted">Waiting for another course to finish…</p>
  if (run.state === 'running')
    return (
      <>
        <p className="flex items-baseline justify-between gap-3">
          <span className="text-[15px] font-semibold">
            Translating {Math.min(run.finished + 1, run.total)}/{run.total}
          </span>
          <span className="text-muted tabular-nums">
            {completionPercent(run.finished, run.total)}%
          </span>
        </p>
        <p className="wrap-anywhere text-muted">{run.currentLecture ?? '…'}</p>
        <ProgressBar className="h-2 rounded" completed={run.finished} total={run.total} />
        <p className="text-xs text-muted">Spent {formatUsd(run.costUsd)}</p>
      </>
    )
  const translated = run.finished - run.failed.length
  return (
    <p>
      {FINISHED_TEXT[run.state]} {translated}/{run.total} lectures translated.{' '}
      <span className="text-muted">Spent {formatUsd(run.costUsd)}.</span>
    </p>
  )
}
