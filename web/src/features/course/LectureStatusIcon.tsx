import type { ReactElement } from 'react'
import { Check, Circle, CircleDot, type LucideIcon } from 'lucide-react'
import { LECTURE_STATUS_LABELS, LectureStatus } from './lecture-status'

interface LectureStatusIconProps {
  readonly status: LectureStatus
}

const STATUS_ICONS: Readonly<Record<LectureStatus, LucideIcon>> = {
  [LectureStatus.Completed]: Check,
  [LectureStatus.InProgress]: CircleDot,
  [LectureStatus.NotStarted]: Circle,
}

export function LectureStatusIcon({ status }: LectureStatusIconProps): ReactElement {
  const Icon = STATUS_ICONS[status]
  return (
    <Icon
      className="size-3.5 shrink-0 text-accent-soft"
      strokeWidth={status === LectureStatus.Completed ? 3 : 2}
      role="img"
      aria-label={LECTURE_STATUS_LABELS[status]}
    />
  )
}
