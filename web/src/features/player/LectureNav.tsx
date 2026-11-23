import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { lecturePath } from '../../app/paths'
import type { LectureNeighbours } from '../course/course-navigation'
import { ChevronIcon } from '@videojs/react/icons'

interface LectureNavProps {
  readonly courseId: string
  readonly neighbours: LectureNeighbours
}

export function LectureNav({ courseId, neighbours }: LectureNavProps): ReactElement {
  const { previous, next } = neighbours
  return (
    <nav
      className="flex justify-between gap-4 border-t border-line py-3 text-accent"
      aria-label="Chuyển bài"
    >
      {previous !== undefined ? (
        <Link
          className="inline-flex items-center gap-1 wrap-anywhere"
          to={lecturePath(courseId, previous.id)}
        >
          <ChevronIcon className="size-4.5 -scale-x-100" />
          {previous.title}
        </Link>
      ) : (
        <span />
      )}
      {next !== undefined && (
        <Link
          className="inline-flex items-center gap-1 text-right wrap-anywhere"
          to={lecturePath(courseId, next.id)}
        >
          {next.title}
          <ChevronIcon className="size-4.5" />
        </Link>
      )}
    </nav>
  )
}
