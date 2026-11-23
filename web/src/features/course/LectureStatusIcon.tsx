import { CheckIcon } from '@videojs/react/icons'
import type { ReactElement } from 'react'
import { LECTURE_STATUS_LABELS, LectureStatus } from './lecture-status'

interface LectureStatusIconProps {
  readonly status: LectureStatus
}

const DOT = 'inline-block size-3 rounded-full border-2'

export function LectureStatusIcon({ status }: LectureStatusIconProps): ReactElement {
  return (
    <span
      className="inline-grid size-4.5 place-items-center text-accent"
      role="img"
      aria-label={LECTURE_STATUS_LABELS[status]}
    >
      {status === LectureStatus.Completed && <CheckIcon className="size-4.5" />}
      {status === LectureStatus.InProgress && (
        <span
          className={`${DOT} border-accent bg-[linear-gradient(90deg,var(--accent)_50%,transparent_50%)]`}
        />
      )}
      {status === LectureStatus.NotStarted && <span className={`${DOT} border-line`} />}
    </span>
  )
}
