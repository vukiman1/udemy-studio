import { useEffect, useRef, type ReactElement } from 'react'
import { NavLink, useParams } from 'react-router'
import { ChevronDown, List } from 'lucide-react'
import type { CourseDetail, LectureSummary } from '../../api/types'
import { lecturePath } from '../../app/paths'
import { lectureStatus } from './lecture-status'
import { LectureStatusIcon } from './LectureStatusIcon'

interface LectureSidebarProps {
  readonly course: CourseDetail
}

const LECTURE_NUMBER_DIGITS = 3
const CHAPTER_NUMBER_DIGITS = 2

export function LectureSidebar({ course }: LectureSidebarProps): ReactElement {
  const { lectureId } = useParams()
  const activeRef = useRef<HTMLAnchorElement | null>(null)

  useEffect(() => {
    const activeLink = activeRef.current
    if (activeLink === null) return
    const chapter = activeLink.closest('details')
    if (chapter !== null) chapter.open = true
    activeLink.scrollIntoView({ block: 'nearest' })
  }, [lectureId])

  return (
    <nav className="grid gap-3.5" aria-labelledby="lecture-list-title">
      <h2
        id="lecture-list-title"
        className="sticky top-0 z-10 flex items-center gap-3 bg-sidebar px-2 py-4 text-lg font-semibold max-lg:static"
      >
        <List className="size-5" aria-hidden="true" />
        Nội dung
      </h2>
      {course.chapters.map((chapter) => (
        <details
          key={`${String(chapter.index)}-${chapter.title}`}
          className="group overflow-hidden bg-card"
          open
        >
          <summary className="flex cursor-pointer list-none items-center gap-2.5 bg-card-header py-2.5 pr-3 pl-3.5 text-[13px] font-bold tracking-[0.02em] uppercase [&::-webkit-details-marker]:hidden">
            <span className="flex-1">
              {String(chapter.index).padStart(CHAPTER_NUMBER_DIGITS, '0')} · {chapter.title}
            </span>
            <ChevronDown
              className="size-4 text-muted transition-transform group-not-open:-rotate-90"
              aria-hidden="true"
            />
          </summary>
          <ol className="list-none">
            {chapter.lectures.map((lecture) => (
              <li key={lecture.id}>
                <NavLink
                  ref={lecture.id === lectureId ? activeRef : undefined}
                  to={lecturePath(course.id, lecture.id)}
                  className="group/row flex items-center gap-2.5 px-4 py-[13px] text-[13.5px] hover:bg-ink/5 aria-[current=page]:bg-selected aria-[current=page]:font-semibold"
                >
                  <LectureStatusIcon status={lectureStatus(lecture)} />
                  <span className="w-[30px] shrink-0 text-[13px] font-medium text-muted tabular-nums group-aria-[current=page]/row:text-ink">
                    {String(lecture.number).padStart(LECTURE_NUMBER_DIGITS, '0')}
                  </span>
                  <span className="flex-1 wrap-anywhere">{lecture.title}</span>
                  {lecture.kind !== 'video' && (
                    <span className="shrink-0 text-xs font-normal text-muted">
                      {KIND_LABELS[lecture.kind]}
                    </span>
                  )}
                </NavLink>
              </li>
            ))}
          </ol>
        </details>
      ))}
    </nav>
  )
}

const KIND_LABELS = {
  article: 'Bài viết',
  quiz: 'Quiz',
  unavailable: 'DRM',
} as const satisfies Record<Exclude<LectureSummary['kind'], 'video'>, string>
