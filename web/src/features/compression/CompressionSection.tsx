import type { ReactElement } from 'react'
import { COMPRESSION_UNAVAILABLE_STATUS, useCompressionRun } from '../../api/compression-queries'
import { ApiRequestError, useCourses } from '../../api/queries'
import type { CourseSummary } from '../../api/types'
import { PageStatus } from '../../app/PageStatus'
import { CompressionCard } from './CompressionCard'

export function CompressionSection(): ReactElement {
  const courses = useCourses()
  return (
    <section className="grid gap-3">
      <div>
        <h2 className="font-semibold">Nén video (H.265)</h2>
        <p className="text-[0.9rem] text-muted">
          Nén lại các video H.264 cho nhẹ hơn, chạy nền khi máy rảnh. Chỉ thay file gốc khi bản nén
          dài bằng và nhỏ hơn bản gốc.
        </p>
      </div>
      {courses.isError && <PageStatus tone="error" message={courses.error.message} />}
      {courses.data?.length === 0 && <p className="text-muted">Thư viện chưa có khoá nào.</p>}
      {courses.data !== undefined && courses.data.length > 0 && (
        <CompressionCards courses={courses.data} />
      )}
    </section>
  )
}

interface CompressionCardsProps {
  readonly courses: readonly CourseSummary[]
}

// Availability comes from the first course's status, a query its card shares, so a missing ffmpeg is said once.
function CompressionCards({ courses }: CompressionCardsProps): ReactElement {
  const firstRun = useCompressionRun(courses[0]?.id ?? '')
  if (
    firstRun.error instanceof ApiRequestError &&
    firstRun.error.status === COMPRESSION_UNAVAILABLE_STATUS
  )
    return <PageStatus tone="error" message={firstRun.error.message} />
  return (
    <>
      {courses.map((course) => (
        <CompressionCard key={course.id} courseId={course.id} title={course.title} />
      ))}
    </>
  )
}
