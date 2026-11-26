import type { UseQueryResult } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { Cloud } from 'lucide-react'
import { Link } from 'react-router'
import type { PublishedCourse } from '../../api/types'
import { coursePath } from '../../app/paths'
import { formatBytes } from '../compression/format'

const PUBLISHED_AT_FORMAT = new Intl.DateTimeFormat('vi-VN', {
  dateStyle: 'short',
  timeStyle: 'short',
})

interface PublishedCourseBannerProps {
  readonly published: UseQueryResult<PublishedCourse | null>
}

export function PublishedCourseBanner({ published }: PublishedCourseBannerProps): ReactElement {
  return (
    <section
      className="flex flex-wrap items-center gap-x-3 gap-y-1 rounded-2xl bg-card px-5 py-3.5 text-[13px]"
      aria-label="Khoá học trên R2"
    >
      <span className="flex items-center gap-2 font-semibold text-accent-soft">
        <Cloud className="size-[18px]" aria-hidden="true" />
        Trên R2
      </span>
      <BannerBody published={published} />
    </section>
  )
}

function BannerBody({ published }: PublishedCourseBannerProps): ReactElement {
  if (published.isError) return <span className="text-danger">{published.error.message}</span>
  if (published.data === undefined) return <span className="text-muted">…</span>
  if (published.data === null) return <span className="text-muted">Trống</span>
  const course = published.data
  return (
    <>
      {course.courseId === undefined ? (
        <span className="font-semibold">{course.title}</span>
      ) : (
        <Link className="font-semibold hover:underline" to={coursePath(course.courseId)}>
          {course.title}
        </Link>
      )}
      <span className="text-muted">
        {isPartial(course) &&
          `${String(course.chapters)}/${String(course.totalChapters)} chương · `}
        {formatBytes(course.sizeBytes)} · {PUBLISHED_AT_FORMAT.format(new Date(course.publishedAt))}
        {course.courseId === undefined && ' · không có trên máy'}
      </span>
    </>
  )
}

function isPartial(course: PublishedCourse): boolean {
  return course.totalChapters > 0 && course.chapters < course.totalChapters
}
