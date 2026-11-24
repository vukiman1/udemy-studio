import type { ReactElement, RefObject } from 'react'
import { courseNotesUrl, useLectureNotes } from '../../api/note-queries'
import type { LectureSummary } from '../../api/types'
import { NoteComposer } from './NoteComposer'
import { NoteItem } from './NoteItem'

interface NotesPanelProps {
  readonly lecture: LectureSummary
  readonly courseId: string
  // Null for lectures without video: their notes have no moment.
  readonly videoRef: RefObject<HTMLVideoElement | null> | null
}

export function NotesPanel({ lecture, courseId, videoRef }: NotesPanelProps): ReactElement {
  const notes = useLectureNotes(lecture.id)
  const seekTo =
    videoRef === null
      ? undefined
      : (seconds: number): void => {
          if (videoRef.current !== null) videoRef.current.currentTime = seconds
        }

  return (
    <section
      className="grid gap-3 rounded-xl border border-line bg-surface p-4"
      aria-label="Ghi chú"
    >
      <header className="flex flex-wrap items-center gap-3">
        <h3 className="font-semibold">Ghi chú</h3>
        <a
          href={courseNotesUrl(courseId)}
          download
          className="ml-auto text-[0.9rem] text-accent underline"
        >
          Tải ghi chú cả khoá (.md)
        </a>
      </header>
      <NoteComposer key={lecture.id} lectureId={lecture.id} videoRef={videoRef} />
      {notes.isError && (
        <p className="m-0 text-danger" role="alert">
          Không tải được ghi chú ({notes.error.message}).
        </p>
      )}
      {notes.data !== undefined && notes.data.length > 0 && (
        <ol className="grid gap-2">
          {notes.data.map((note) => (
            <NoteItem key={note.id} note={note} lectureId={lecture.id} onSeek={seekTo} />
          ))}
        </ol>
      )}
    </section>
  )
}
