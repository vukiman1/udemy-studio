import { Book, Download, WifiOff } from 'lucide-react'
import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { useCourseTools } from '../../api/library-tools-queries'
import { ApiRequestError, useCourses } from '../../api/queries'
import { useIsCloudCopy, useIsLibraryCopy } from '../../api/server-queries'
import { DOWNLOADS_PATH } from '../../app/paths'
import { StaleDataNotice } from '../../app/StaleDataNotice'
import { Button, ButtonSize, ButtonVariant, buttonClass } from '../../ui/Button'
import { EmptyState } from '../../ui/EmptyState'
import { CourseCard } from './CourseCard'
import { PublishedCourseBanner } from './PublishedCourseBanner'
import { RefreshLibraryButton } from './RefreshLibraryButton'
import { useCoursePublishing } from './use-course-publishing'

const SKELETON_ROWS = ['first', 'second', 'third'] as const

export function LibraryPage(): ReactElement {
  const courses = useCourses()
  const isLibraryCopy = useIsLibraryCopy()
  const isCloudCopy = useIsCloudCopy()
  const tools = useCourseTools(isLibraryCopy)
  const publishing = useCoursePublishing(isLibraryCopy)

  return (
    <main className="grid gap-7 px-5 py-8 md:px-10 md:py-10">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="m-0 text-[30px] font-semibold">Library</h1>
        {isCloudCopy && <RefreshLibraryButton />}
      </div>
      {publishing.isConfigured && <PublishedCourseBanner published={publishing.published} />}
      {courses.data === undefined ? (
        courses.isError ? (
          <LoadFailure
            error={courses.error}
            onRetry={() => {
              void courses.refetch()
            }}
          />
        ) : (
          <LibrarySkeleton />
        )
      ) : (
        <>
          {courses.isError && <StaleDataNotice error={courses.error} />}
          {courses.data.length === 0 ? (
            <EmptyState
              icon={Book}
              title="No courses yet"
              body="Download a course to start."
              action={
                <Link
                  className={buttonClass(ButtonVariant.Primary, ButtonSize.Medium)}
                  to={DOWNLOADS_PATH}
                >
                  <Download aria-hidden="true" />
                  Download a course
                </Link>
              }
            />
          ) : (
            <ul className="stagger-children m-0 grid list-none gap-[22px] p-0">
              {courses.data.map((course) => (
                <li key={course.id}>
                  <CourseCard
                    course={course}
                    tools={tools.data?.find((entry) => entry.courseId === course.id)}
                    publish={publishing.of(course.id)}
                  />
                </li>
              ))}
            </ul>
          )}
        </>
      )}
    </main>
  )
}

interface LoadFailureProps {
  readonly error: Error
  readonly onRetry: () => void
}

function LoadFailure({ error, onRetry }: LoadFailureProps): ReactElement {
  const isUnreachable = !(error instanceof ApiRequestError)
  return (
    <EmptyState
      icon={WifiOff}
      isError
      title={isUnreachable ? 'Cannot reach udemy-studio' : 'Could not load the library'}
      body={isUnreachable ? 'Check the server is running, then reload.' : error.message}
      action={
        <Button variant={ButtonVariant.Primary} onClick={onRetry}>
          Reload
        </Button>
      }
    />
  )
}

function LibrarySkeleton(): ReactElement {
  return (
    <ul className="m-0 grid list-none gap-[22px] p-0" aria-label="Loading library…">
      {SKELETON_ROWS.map((row) => (
        <li key={row} className="flex animate-pulse items-center gap-5 rounded-2xl bg-card p-3">
          <div className="aspect-video w-24 rounded-[10px] bg-surface-raised md:w-40" />
          <div className="grid flex-1 gap-3">
            <div className="h-4 w-2/5 rounded bg-surface-raised" />
            <div className="h-2 rounded bg-surface-raised" />
            <div className="h-3 w-1/4 rounded bg-surface-raised" />
          </div>
        </li>
      ))}
    </ul>
  )
}
