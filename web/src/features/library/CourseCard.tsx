import type { ReactElement } from 'react'
import { Link } from 'react-router'
import type { CourseSummary } from '../../api/types'
import { coursePath } from '../../app/paths'
import { ProgressBar } from '../../app/ProgressBar'
import styles from './CourseCard.module.css'

interface CourseCardProps {
  readonly course: CourseSummary
}

export function CourseCard({ course }: CourseCardProps): ReactElement {
  return (
    <Link className={styles.card} to={coursePath(course.id)}>
      <h2 className={styles.title}>{course.title}</h2>
      <ProgressBar completed={course.completedLectures} total={course.playableLectures} />
      <span className={styles.meta}>
        {course.completedLectures}/{course.playableLectures} bài đã học
      </span>
    </Link>
  )
}
