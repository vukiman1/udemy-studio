import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { lecturePath } from '../../app/paths'
import type { LectureNeighbours } from '../course/course-navigation'

interface LectureNavProps {
  readonly courseId: string
  readonly neighbours: LectureNeighbours
}

const NEIGHBOUR_LINK =
  'inline-flex items-center gap-1.5 wrap-break-word [&_svg]:size-[18px] [&_svg]:shrink-0'

export function LectureNav({ courseId, neighbours }: LectureNavProps): ReactElement {
  const { previous, next } = neighbours
  return (
    <nav
      className="flex justify-between gap-4 border-t border-line pt-4 text-[15px] font-medium text-accent-soft"
      aria-label="Lecture navigation"
    >
      {previous !== undefined ? (
        <Link className={NEIGHBOUR_LINK} to={lecturePath(courseId, previous.id)}>
          <ChevronLeft aria-hidden="true" />
          {previous.title}
        </Link>
      ) : (
        <span />
      )}
      {next !== undefined && (
        <Link className={`${NEIGHBOUR_LINK} text-right`} to={lecturePath(courseId, next.id)}>
          {next.title}
          <ChevronRight aria-hidden="true" />
        </Link>
      )}
    </nav>
  )
}
