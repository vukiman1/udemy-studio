import type { ReactElement } from 'react'
import { Link, Outlet, useParams } from 'react-router'
import { useCourse } from '../../api/queries'
import { PageStatus } from '../../app/PageStatus'
import { StaleDataNotice } from '../../app/StaleDataNotice'
import { ProgressBar } from '../../app/ProgressBar'
import { LectureSidebar } from './LectureSidebar'
import styles from './CourseLayout.module.css'

export function CourseLayout(): ReactElement {
  const { courseId = '' } = useParams()
  const course = useCourse(courseId)

  if (course.data === undefined) {
    return course.isError ? (
      <PageStatus tone="error" message={`Không tải được khoá học: ${course.error.message}`} />
    ) : (
      <PageStatus tone="info" message="Đang tải khoá học…" />
    )
  }

  return (
    <div className={styles.layout}>
      <header className={styles.header}>
        <Link to="/" className={styles.back}>
          ← Thư viện
        </Link>
        <h1 className={styles.title}>{course.data.title}</h1>
        <div className={styles.progress}>
          <ProgressBar
            completed={course.data.completedLectures}
            total={course.data.playableLectures}
          />
        </div>
      </header>
      <main className={styles.main}>
        {course.isError && <StaleDataNotice error={course.error} />}
        <Outlet />
      </main>
      <aside className={styles.sidebar}>
        <LectureSidebar course={course.data} />
      </aside>
    </div>
  )
}
