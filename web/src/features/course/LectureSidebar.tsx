import { useEffect, useRef, type ReactElement } from 'react'
import { NavLink, useParams } from 'react-router'
import type { CourseDetail, LectureSummary } from '../../api/types'
import { lecturePath } from '../../app/paths'
import { lectureStatus } from './lecture-status'
import { LectureStatusIcon } from './LectureStatusIcon'

interface LectureSidebarProps {
  readonly course: CourseDetail
}

const LECTURE_NUMBER_DIGITS = 3

export function LectureSidebar({ course }: LectureSidebarProps): ReactElement {
  const { lectureId } = useParams()
  const activeRef = useRef<HTMLAnchorElement | null>(null)

  useEffect(() => {
    activeRef.current?.scrollIntoView({ block: 'nearest' })
  }, [lectureId])

  return (
    <nav aria-label="Danh sách bài học">
      {course.chapters.map((chapter) => (
        <section key={`${String(chapter.index)}-${chapter.title}`} className="py-2">
          <h2 className="mx-4 my-2 text-[0.8rem] tracking-[0.04em] text-muted uppercase">
            {String(chapter.index).padStart(2, '0')} · {chapter.title}
          </h2>
          <ol className="list-none">
            {chapter.lectures.map((lecture) => (
              <li key={lecture.id}>
                <NavLink
                  ref={lecture.id === lectureId ? activeRef : undefined}
                  to={lecturePath(course.id, lecture.id)}
                  className="grid grid-cols-[1.2em_2.4em_1fr_auto] items-center gap-1.5 px-4 py-2 text-[0.92rem] hover:bg-surface aria-[current=page]:bg-surface-strong aria-[current=page]:font-semibold"
                >
                  <LectureStatusIcon status={lectureStatus(lecture)} />
                  <span className="text-muted tabular-nums">
                    {String(lecture.number).padStart(LECTURE_NUMBER_DIGITS, '0')}
                  </span>
                  <span className="wrap-anywhere">{lecture.title}</span>
                  {lecture.kind !== 'video' && (
                    <span className="rounded border border-line px-1 text-[0.75rem] text-muted">
                      {kindLabel(lecture)}
                    </span>
                  )}
                </NavLink>
              </li>
            ))}
          </ol>
        </section>
      ))}
    </nav>
  )
}

function kindLabel(lecture: LectureSummary): string {
  return lecture.kind === 'article' ? 'Bài viết' : 'DRM'
}
