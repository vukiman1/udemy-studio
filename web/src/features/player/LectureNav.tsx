import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { lecturePath } from '../../app/paths'
import type { LectureNeighbours } from '../course/course-navigation'
import styles from './LectureNav.module.css'

interface LectureNavProps {
  readonly courseId: string
  readonly neighbours: LectureNeighbours
}

export function LectureNav({ courseId, neighbours }: LectureNavProps): ReactElement {
  const { previous, next } = neighbours
  return (
    <nav className={styles.nav} aria-label="Chuyển bài">
      {previous !== undefined ? (
        <Link to={lecturePath(courseId, previous.id)}>← {previous.title}</Link>
      ) : (
        <span />
      )}
      {next !== undefined && (
        <Link className={styles.next} to={lecturePath(courseId, next.id)}>
          {next.title} →
        </Link>
      )}
    </nav>
  )
}
