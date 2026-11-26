import { Archive, Cloud, CloudUpload, Languages, Play } from 'lucide-react'
import { useState, type ReactElement } from 'react'
import { Link } from 'react-router'
import type { CourseSummary, CourseTools, PublishRun } from '../../api/types'
import { coursePath, lecturePath } from '../../app/paths'
import { ProgressBar } from '../../app/ProgressBar'
import { ButtonSize, ButtonVariant, buttonClass } from '../../ui/Button'
import { CompressionDialog } from '../compression/CompressionDialog'
import { formatBytes } from '../compression/format'
import { PublishDialog } from '../publish/PublishDialog'
import { TranslationDialog } from '../translation/TranslationDialog'
import { CourseThumb } from './CourseThumb'
import { CourseToolButton } from './CourseToolButton'
import { ToolState, compressionStatus, publishStatus, translationStatus } from './tool-state'

export interface CoursePublish {
  readonly run: PublishRun | undefined
  readonly isPublished: boolean
}

interface CourseCardProps {
  readonly course: CourseSummary
  readonly tools: CourseTools | undefined
  readonly publish: CoursePublish | undefined
}

enum OpenTool {
  Translation = 'translation',
  Compression = 'compression',
  Publish = 'publish',
}

const PERCENT = 100

export function CourseCard({ course, tools, publish }: CourseCardProps): ReactElement {
  const [openTool, setOpenTool] = useState<OpenTool | undefined>(undefined)
  const translation = tools === undefined ? undefined : translationStatus(tools.translation)
  const compression = tools === undefined ? undefined : compressionStatus(tools.compression)
  const publishing =
    publish === undefined ? undefined : publishStatus(publish.run, publish.isPublished)
  const running = [translation, compression, publishing].find(
    (status) => status?.state === ToolState.Running,
  )
  const resumeTo =
    course.resume === undefined
      ? coursePath(course.id)
      : lecturePath(course.resume.courseId, course.resume.lectureId)
  const close = (): void => {
    setOpenTool(undefined)
  }

  return (
    <article className="group/card relative flex flex-wrap items-start gap-x-5 gap-y-3 overflow-hidden rounded-2xl bg-card p-3 transition-[translate] duration-300 ease-soft motion-safe:hover:-translate-y-1 md:flex-nowrap md:items-center md:pr-5">
      <Link to={coursePath(course.id)} tabIndex={-1} aria-hidden="true">
        <CourseThumb
          course={course}
          className="block aspect-video w-24 rounded-[10px] transition-[scale] duration-300 ease-soft motion-safe:group-hover/card:scale-[1.04] md:w-40"
        />
      </Link>
      <div className="grid min-w-0 flex-1 gap-3.5 py-1.5">
        <h2 className="m-0 text-[17px] leading-snug font-semibold wrap-anywhere">
          <Link to={coursePath(course.id)}>{course.title}</Link>
        </h2>
        <div className="grid gap-2">
          <ProgressBar
            completed={course.completedLectures}
            total={course.playableLectures}
            className="h-2 rounded"
          />
          <span className="flex flex-wrap items-center gap-x-2 text-[13px] text-muted">
            {course.completedLectures} / {course.playableLectures} bài đã học ·{' '}
            {formatBytes(course.sizeBytes)}
            {publish?.isPublished === true && (
              <span className="inline-flex items-center gap-1 rounded-full bg-accent/15 px-2 py-0.5 text-xs font-semibold text-accent-soft">
                <Cloud className="size-3.5" aria-hidden="true" />
                R2
              </span>
            )}
          </span>
        </div>
      </div>
      <div className="flex basis-full flex-wrap gap-2 md:basis-auto md:flex-nowrap">
        <Link className={buttonClass(ButtonVariant.Primary, ButtonSize.Medium)} to={resumeTo}>
          <Play aria-hidden="true" />
          Tiếp tục
        </Link>
        {translation !== undefined && (
          <CourseToolButton
            icon={Languages}
            status={translation}
            onOpen={() => {
              setOpenTool(OpenTool.Translation)
            }}
          />
        )}
        {compression !== undefined && compression.state !== ToolState.Unavailable && (
          <CourseToolButton
            icon={Archive}
            status={compression}
            onOpen={() => {
              setOpenTool(OpenTool.Compression)
            }}
          />
        )}
        {publishing !== undefined && (
          <CourseToolButton
            icon={CloudUpload}
            status={publishing}
            onOpen={() => {
              setOpenTool(OpenTool.Publish)
            }}
          />
        )}
      </div>
      {running !== undefined && (
        <div
          className="absolute bottom-0 left-0 h-[3px] bg-accent"
          style={{ width: `${String(Math.round(running.progress * PERCENT))}%` }}
          aria-hidden="true"
        />
      )}
      {openTool === OpenTool.Translation && tools !== undefined && (
        <TranslationDialog
          courseId={course.id}
          title={course.title}
          run={tools.translation.run}
          onClose={close}
        />
      )}
      {openTool === OpenTool.Compression && tools !== undefined && (
        <CompressionDialog
          courseId={course.id}
          title={course.title}
          run={tools.compression?.run}
          onClose={close}
        />
      )}
      {openTool === OpenTool.Publish && publish !== undefined && (
        <PublishDialog
          courseId={course.id}
          title={course.title}
          run={publish.run}
          onClose={close}
        />
      )}
    </article>
  )
}
