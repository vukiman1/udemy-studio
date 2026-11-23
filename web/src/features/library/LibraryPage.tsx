import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { useContinueTarget, useCourses } from '../../api/queries'
import type { CourseSummary } from '../../api/types'
import { PageStatus } from '../../app/PageStatus'
import { StaleDataNotice } from '../../app/StaleDataNotice'
import { lecturePath } from '../../app/paths'
import { CourseCard } from './CourseCard'
import styles from './LibraryPage.module.css'

export function LibraryPage(): ReactElement {
  const courses = useCourses()
  const continueTarget = useContinueTarget()

  if (courses.data === undefined) {
    return courses.isError ? (
      <PageStatus tone="error" message={`Không tải được thư viện: ${courses.error.message}`} />
    ) : (
      <PageStatus tone="info" message="Đang tải thư viện…" />
    )
  }

  const target = continueTarget.data ?? null
  const resumeCourse =
    target === null ? undefined : courses.data.find((course) => course.id === target.courseId)

  return (
    <main className={styles.page}>
      <header className={styles.header}>
        <h1>Thư viện</h1>
        {target !== null && resumeCourse !== undefined && (
          <Link className={styles.continue} to={lecturePath(target.courseId, target.lectureId)}>
            ▶ Tiếp tục học · {resumeCourse.title}
          </Link>
        )}
      </header>
      {courses.isError && <StaleDataNotice error={courses.error} />}
      {courses.data.length === 0 ? (
        <PageStatus tone="info" message="Chưa có khoá học nào trong thư mục downloads/." />
      ) : (
        <ul className={styles.grid}>
          {courses.data.map((course: CourseSummary) => (
            <li key={course.id}>
              <CourseCard course={course} />
            </li>
          ))}
        </ul>
      )}
    </main>
  )
}
