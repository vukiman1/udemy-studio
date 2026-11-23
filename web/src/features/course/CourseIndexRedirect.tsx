import type { ReactElement } from 'react'
import { Navigate, useParams } from 'react-router'
import { useCourse } from '../../api/queries'
import { PageStatus } from '../../app/PageStatus'
import { lecturePath } from '../../app/paths'
import { startingLectureId } from './course-navigation'

export function CourseIndexRedirect(): ReactElement {
  const { courseId = '' } = useParams()
  const course = useCourse(courseId)
  if (course.data === undefined) return <PageStatus tone="info" message="Đang mở khoá học…" />
  const lectureId = startingLectureId(course.data)
  if (lectureId === undefined)
    return <PageStatus tone="info" message="Khoá học này chưa có bài nào xem được." />
  return <Navigate replace to={lecturePath(courseId, lectureId)} />
}
