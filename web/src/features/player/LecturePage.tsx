import { useRef, type ReactElement } from 'react'
import { useParams, useSearchParams } from 'react-router'
import { Info } from 'lucide-react'
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
    <article
      key={lecture.id}
      className="grid content-start gap-4 px-5 pt-6 pb-10 motion-safe:animate-page-enter md:px-10"
    >
      <h2 className="text-xl font-semibold">{lecture.title}</h2>
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
        <p className="flex items-center gap-2.5 rounded-2xl bg-card p-5 text-sm" role="status">
          <Info className="size-4 shrink-0 text-muted" aria-hidden="true" />
          Bài này có DRM nên không tải được video; chỉ có phụ đề hoặc tài liệu kèm theo.
        </p>
      )}
      {needsManualCompletion(lecture) && (
        <MarkCompleteButton lecture={lecture} courseId={courseId} />
      )}
      <LectureNav courseId={courseId} neighbours={neighboursOf(course.data, lecture.id)} />
      <AttachmentList attachments={lecture.attachments} />
      <NotesPanel
        lecture={lecture}
        courseId={courseId}
        videoRef={lecture.hasVideo ? videoRef : null}
      />
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
