import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { useContinueTarget, useCourses } from '../../api/queries'
import type { CourseSummary } from '../../api/types'
import { PageStatus } from '../../app/PageStatus'
import { StaleDataNotice } from '../../app/StaleDataNotice'
import { lecturePath } from '../../app/paths'
import { CourseCard } from './CourseCard'
import { PlayIcon } from '@videojs/react/icons'

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
    <main className="mx-auto max-w-[1100px] px-4 py-6">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-2xl font-bold">Thư viện</h1>
        {target !== null && resumeCourse !== undefined && (
          <Link
            className="inline-flex items-center gap-2 rounded-lg bg-accent px-4 py-2.5 font-semibold text-on-accent"
            to={lecturePath(target.courseId, target.lectureId)}
          >
            <PlayIcon className="size-4.5" />
            Tiếp tục học · {resumeCourse.title}
          </Link>
        )}
      </header>
      {courses.isError && <StaleDataNotice error={courses.error} />}
      {courses.data.length === 0 ? (
        <PageStatus tone="info" message="Chưa có khoá học nào trong thư mục downloads/." />
      ) : (
        <ul className="mt-6 grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-4">
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
