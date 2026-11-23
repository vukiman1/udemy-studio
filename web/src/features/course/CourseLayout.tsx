import type { ReactElement } from 'react'
import { Link, Outlet, useParams } from 'react-router'
import { useCourse } from '../../api/queries'
import { PageStatus } from '../../app/PageStatus'
import { StaleDataNotice } from '../../app/StaleDataNotice'
import { ProgressBar } from '../../app/ProgressBar'
import { LectureSidebar } from './LectureSidebar'
import { ChevronIcon } from '@videojs/react/icons'

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
    <div className="grid h-dvh grid-cols-[minmax(0,1fr)_340px] grid-rows-[auto_1fr] max-lg:h-auto max-lg:grid-cols-1 max-lg:grid-rows-none">
      <header className="col-span-full flex items-center gap-4 border-b border-line px-4 py-2.5">
        <Link to="/" className="inline-flex items-center gap-1 whitespace-nowrap text-muted">
          <ChevronIcon className="size-4.5 -scale-x-100" />
          Thư viện
        </Link>
        <h1 className="truncate text-base font-semibold">{course.data.title}</h1>
        <div className="ml-auto w-40 shrink-0 max-lg:hidden">
          <ProgressBar
            completed={course.data.completedLectures}
            total={course.data.playableLectures}
          />
        </div>
      </header>
      <main className="overflow-y-auto max-lg:overflow-visible">
        {course.isError && <StaleDataNotice error={course.error} />}
        <Outlet />
      </main>
      <aside className="overflow-y-auto border-l border-line max-lg:overflow-visible max-lg:border-t max-lg:border-l-0">
        <LectureSidebar course={course.data} />
      </aside>
    </div>
  )
}
