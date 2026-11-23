import type { ReactElement } from 'react'
import { useParams } from 'react-router'
import { useCourse } from '../../api/queries'
import type { LectureSummary } from '../../api/types'
import { PageStatus } from '../../app/PageStatus'
import { findLecture, neighboursOf } from '../course/course-navigation'
import { ArticleFrame } from './ArticleFrame'
import { AttachmentList } from './AttachmentList'
import { LectureNav } from './LectureNav'
import { MarkCompleteButton } from './MarkCompleteButton'
import { VideoPlayer } from './VideoPlayer'
import styles from './LecturePage.module.css'

export function LecturePage(): ReactElement {
  const { courseId = '', lectureId = '' } = useParams()
  const course = useCourse(courseId)
  if (course.data === undefined) return <PageStatus tone="info" message="Đang tải bài học…" />
  const lecture = findLecture(course.data, lectureId)
  if (lecture === undefined)
    return <PageStatus tone="error" message="Không tìm thấy bài học này trong khoá." />

  return (
    <article className={styles.page}>
      <h2 className={styles.title}>{lecture.title}</h2>
      {lecture.hasVideo && <VideoPlayer key={lecture.id} lecture={lecture} courseId={courseId} />}
      {lecture.hasArticle && <ArticleFrame lectureId={lecture.id} title={lecture.title} />}
      {lecture.kind === 'unavailable' && (
        <PageStatus
          tone="info"
          message="Bài này có DRM nên không tải được video; chỉ có phụ đề hoặc tài liệu kèm theo."
        />
      )}
      {needsManualCompletion(lecture) && (
        <MarkCompleteButton lecture={lecture} courseId={courseId} />
      )}
      <AttachmentList attachments={lecture.attachments} />
      <LectureNav courseId={courseId} neighbours={neighboursOf(course.data, lecture.id)} />
    </article>
  )
}

function needsManualCompletion(lecture: LectureSummary): boolean {
  return lecture.kind === 'article' && !lecture.hasVideo
}
