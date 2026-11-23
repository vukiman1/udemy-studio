import { useEffect, useRef, type ReactElement } from 'react'
import { NavLink, useParams } from 'react-router'
import type { CourseDetail, LectureSummary } from '../../api/types'
import { lecturePath } from '../../app/paths'
import styles from './LectureSidebar.module.css'

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
        <section key={`${String(chapter.index)}-${chapter.title}`} className={styles.chapter}>
          <h2 className={styles.chapterTitle}>
            {String(chapter.index).padStart(2, '0')} · {chapter.title}
          </h2>
          <ol className={styles.lectures}>
            {chapter.lectures.map((lecture) => (
              <li key={lecture.id}>
                <NavLink
                  ref={lecture.id === lectureId ? activeRef : undefined}
                  to={lecturePath(course.id, lecture.id)}
                  className={({ isActive }) => (isActive ? styles.activeLecture : styles.lecture)}
                >
                  <span className={styles.status} aria-label={statusLabel(lecture)}>
                    {statusIcon(lecture)}
                  </span>
                  <span className={styles.number}>
                    {String(lecture.number).padStart(LECTURE_NUMBER_DIGITS, '0')}
                  </span>
                  <span className={styles.lectureTitle}>{lecture.title}</span>
                  {lecture.kind !== 'video' && (
                    <span className={styles.kind}>{kindLabel(lecture)}</span>
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

function statusIcon(lecture: LectureSummary): string {
  if (lecture.progress?.completed === true) return '✓'
  if (lecture.progress !== undefined) return '◐'
  return '○'
}

function statusLabel(lecture: LectureSummary): string {
  if (lecture.progress?.completed === true) return 'Đã học'
  if (lecture.progress !== undefined) return 'Đang học'
  return 'Chưa học'
}

function kindLabel(lecture: LectureSummary): string {
  return lecture.kind === 'article' ? 'Bài viết' : 'DRM'
}
