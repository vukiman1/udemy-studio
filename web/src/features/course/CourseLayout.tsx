import type { ReactElement } from 'react'
import { Link, Outlet, useParams } from 'react-router'
import { ChevronLeft } from 'lucide-react'
import { useCourse } from '../../api/queries'
import { PageStatus } from '../../app/PageStatus'
import { StaleDataNotice } from '../../app/StaleDataNotice'
import { ProgressBar, completionPercent } from '../../app/ProgressBar'
import { LIBRARY_PATH } from '../../app/paths'
import { LectureSidebar } from './LectureSidebar'

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

  const { title, completedLectures, playableLectures } = course.data
  return (
    <div className="grid h-dvh grid-cols-[minmax(0,1fr)_340px] grid-rows-[auto_1fr] max-lg:h-auto max-lg:grid-cols-1 max-lg:grid-rows-none">
      <header className="col-span-full flex items-center gap-4 px-5 pt-6 pb-4 text-[15px] font-medium text-breadcrumb">
        <nav className="flex min-w-0 flex-1 items-center gap-3.5" aria-label="Vị trí">
          <Link to={LIBRARY_PATH} className="inline-flex shrink-0 items-center gap-1">
            <ChevronLeft className="size-[18px]" aria-hidden="true" />
            Thư viện
          </Link>
          <span aria-hidden="true">/</span>
          <h1 className="truncate">{title}</h1>
        </nav>
        <div className="flex shrink-0 items-center gap-3 text-[13px] max-lg:hidden">
          <span className="font-normal text-muted">Tiến độ</span>
          <ProgressBar
            className="h-1.5 w-40 rounded-[3px]"
            completed={completedLectures}
            total={playableLectures}
          />
          <span className="text-ink tabular-nums">
            {completedLectures} / {playableLectures} (
            {completionPercent(completedLectures, playableLectures)}%)
          </span>
        </div>
      </header>
      <main className="[scrollbar-width:thin] overflow-y-auto max-lg:overflow-visible">
        {course.isError && <StaleDataNotice error={course.error} />}
        <Outlet />
      </main>
      <aside className="[scrollbar-width:thin] overflow-y-auto border-l border-surface-raised bg-sidebar px-3 py-6 max-lg:overflow-visible max-lg:border-t max-lg:border-l-0">
        <LectureSidebar course={course.data} />
      </aside>
    </div>
  )
}
