import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { lecturePath } from '../../app/paths'
import type { LectureNeighbours } from '../course/course-navigation'

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
        <Link className="wrap-anywhere" to={lecturePath(courseId, previous.id)}>
          ← {previous.title}
        </Link>
      ) : (
        <span />
      )}
      {next !== undefined && (
        <Link className="text-right wrap-anywhere" to={lecturePath(courseId, next.id)}>
          {next.title} →
        </Link>
      )}
    </nav>
  )
}
