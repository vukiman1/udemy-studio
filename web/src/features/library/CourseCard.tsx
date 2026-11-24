import { Languages, Shrink } from 'lucide-react'
import { useState, type ReactElement } from 'react'
import { Link } from 'react-router'
import type { CourseSummary, CourseTools } from '../../api/types'
import { coursePath } from '../../app/paths'
import { ProgressBar } from '../../app/ProgressBar'
import { CompressionDialog } from '../compression/CompressionDialog'
import { formatBytes } from '../compression/format'
import { TranslationDialog } from '../translation/TranslationDialog'
import { CourseToolButton } from './CourseToolButton'
import { ToolState, compressionStatus, translationStatus } from './tool-state'

interface CourseCardProps {
  readonly course: CourseSummary
  readonly tools: CourseTools | undefined
}

enum OpenTool {
  Translation = 'translation',
  Compression = 'compression',
}

const PERCENT = 100

export function CourseCard({ course, tools }: CourseCardProps): ReactElement {
  const [openTool, setOpenTool] = useState<OpenTool | undefined>(undefined)
  const translation = tools === undefined ? undefined : translationStatus(tools.translation)
  const compression = tools === undefined ? undefined : compressionStatus(tools.compression)
  const running = [translation, compression].find((status) => status?.state === ToolState.Running)
  const close = (): void => {
    setOpenTool(undefined)
  }

  return (
    <article className="relative grid gap-2.5 overflow-hidden rounded-xl border border-line bg-surface p-4 hover:border-accent">
      <Link className="grid gap-2.5" to={coursePath(course.id)}>
        <h2 className="text-[1.05rem] font-semibold wrap-anywhere">{course.title}</h2>
        <ProgressBar completed={course.completedLectures} total={course.playableLectures} />
        <span className="text-[0.85rem] text-muted">
          {course.completedLectures}/{course.playableLectures} bài đã học ·{' '}
          {formatBytes(course.sizeBytes)}
        </span>
      </Link>
      {translation !== undefined && compression !== undefined && (
        <div className="flex items-center gap-2">
          <CourseToolButton
            icon={Languages}
            status={translation}
            onOpen={() => {
              setOpenTool(OpenTool.Translation)
            }}
          />
          {compression.state !== ToolState.Unavailable && (
            <CourseToolButton
              icon={Shrink}
              status={compression}
              onOpen={() => {
                setOpenTool(OpenTool.Compression)
              }}
            />
          )}
          {running !== undefined && (
            <span className="text-[0.8rem] text-accent">{running.detail}</span>
          )}
        </div>
      )}
      {running !== undefined && (
        <div className="absolute inset-x-0 bottom-0 h-1 bg-surface-strong" aria-hidden="true">
          <div
            className="h-full bg-accent"
            style={{ width: `${String(Math.round(running.progress * PERCENT))}%` }}
          />
        </div>
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
    </article>
  )
}
