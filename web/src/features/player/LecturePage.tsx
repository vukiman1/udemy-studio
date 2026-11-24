import { useRef, type ReactElement } from 'react'
import { useParams, useSearchParams } from 'react-router'
import { useCourse } from '../../api/queries'
import type { LectureSummary } from '../../api/types'
import { PageStatus } from '../../app/PageStatus'
import { START_SECONDS_PARAM } from '../../app/paths'
import { findLecture, neighboursOf } from '../course/course-navigation'
import { NotesPanel } from '../notes/NotesPanel'
import { ArticleFrame } from './ArticleFrame'
import { AttachmentList } from './AttachmentList'
import { LectureNav } from './LectureNav'
import { MarkCompleteButton } from './MarkCompleteButton'
import { VideoPlayer } from './VideoPlayer'

export function LecturePage(): ReactElement {
  const { courseId = '', lectureId = '' } = useParams()
  const [searchParams] = useSearchParams()
  const startAtSeconds = parseStartSeconds(searchParams.get(START_SECONDS_PARAM))
  const course = useCourse(courseId)
  const videoRef = useRef<HTMLVideoElement | null>(null)
  if (course.data === undefined) return <PageStatus tone="info" message="Đang tải bài học…" />
  const lecture = findLecture(course.data, lectureId)
  if (lecture === undefined)
    return <PageStatus tone="error" message="Không tìm thấy bài học này trong khoá." />

  return (
    <article className="mx-auto grid max-w-[1100px] gap-3 p-4">
      <h2 className="text-[1.15rem] font-semibold">{lecture.title}</h2>
      {lecture.hasVideo && (
        <VideoPlayer
          key={`${lecture.id}@${String(startAtSeconds)}`}
          videoRef={videoRef}
          lecture={lecture}
          courseId={courseId}
          startAtSeconds={startAtSeconds}
        />
      )}
      {lecture.hasArticle && (
        <ArticleFrame
          key={lecture.id}
          lectureId={lecture.id}
          title={lecture.title}
          canRunScripts={lecture.kind === 'quiz'}
        />
      )}
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
      <NotesPanel
        lecture={lecture}
        courseId={courseId}
        videoRef={lecture.hasVideo ? videoRef : null}
      />
      <LectureNav courseId={courseId} neighbours={neighboursOf(course.data, lecture.id)} />
    </article>
  )
}

function needsManualCompletion(lecture: LectureSummary): boolean {
  return (lecture.kind === 'article' || lecture.kind === 'quiz') && !lecture.hasVideo
}

function parseStartSeconds(value: string | null): number | undefined {
  if (value === null || value.trim() === '') return undefined
  const seconds = Number(value)
  return Number.isFinite(seconds) && seconds >= 0 ? seconds : undefined
}
