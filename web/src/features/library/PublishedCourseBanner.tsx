import type { UseQueryResult } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { Cloud } from 'lucide-react'
import { Link } from 'react-router'
import type { PublishedCourse, StorageUsage } from '../../api/types'
import { coursePath } from '../../app/paths'
import { StorageUsageMeter } from '../../app/StorageUsageMeter'
import { formatBytes } from '../compression/format'
import { DISPLAY_LOCALE } from '../../app/display-locale'
import { RemovePublishedButton } from './RemovePublishedButton'

const PUBLISHED_AT_FORMAT = new Intl.DateTimeFormat(DISPLAY_LOCALE, {
  dateStyle: 'short',
  timeStyle: 'short',
})

interface PublishedCourseBannerProps {
  readonly published: UseQueryResult<PublishedCourse | null>
  readonly usage: UseQueryResult<StorageUsage | null>
}

export function PublishedCourseBanner({
  published,
  usage,
}: PublishedCourseBannerProps): ReactElement {
  const hasSomethingStored = (usage.data?.usedBytes ?? 0) > 0 || published.data != null
  return (
    <section
      className="flex flex-wrap items-center gap-x-3 gap-y-1 rounded-2xl bg-card px-5 py-3.5 text-[13px]"
      aria-label="Course on R2"
    >
      <span className="flex items-center gap-2 font-semibold text-accent-soft">
        <Cloud className="size-[18px]" aria-hidden="true" />
        On R2
      </span>
      <BannerBody published={published} />
      {usage.data != null && <StorageUsageMeter {...usage.data} />}
      {hasSomethingStored && <RemovePublishedButton />}
    </section>
  )
}

function BannerBody({ published }: Pick<PublishedCourseBannerProps, 'published'>): ReactElement {
  if (published.isError) return <span className="text-danger">{published.error.message}</span>
  if (published.data === undefined) return <span className="text-muted">…</span>
  if (published.data === null) return <span className="text-muted">Empty</span>
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
          `${String(course.chapters)}/${String(course.totalChapters)} chapters · `}
        {formatBytes(course.sizeBytes)} · {PUBLISHED_AT_FORMAT.format(new Date(course.publishedAt))}
        {course.courseId === undefined && ' · not on this machine'}
      </span>
    </>
  )
}

function isPartial(course: PublishedCourse): boolean {
  return course.totalChapters > 0 && course.chapters < course.totalChapters
}
