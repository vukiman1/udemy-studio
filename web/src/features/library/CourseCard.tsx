import type { ReactElement } from 'react'
import { Link } from 'react-router'
import type { CourseSummary } from '../../api/types'
import { coursePath } from '../../app/paths'
import { ProgressBar } from '../../app/ProgressBar'

interface CourseCardProps {
  readonly course: CourseSummary
}

export function CourseCard({ course }: CourseCardProps): ReactElement {
  return (
    <Link
      className="grid gap-2.5 rounded-xl border border-line bg-surface p-4 hover:border-accent"
      to={coursePath(course.id)}
    >
      <h2 className="text-[1.05rem] font-semibold wrap-anywhere">{course.title}</h2>
      <ProgressBar completed={course.completedLectures} total={course.playableLectures} />
      <span className="text-[0.85rem] text-muted">
        {course.completedLectures}/{course.playableLectures} bài đã học
      </span>
    </Link>
  )
}
