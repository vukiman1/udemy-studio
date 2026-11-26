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
    <section className="grid gap-3" aria-label="Notes">
      <NoteComposer key={lecture.id} lectureId={lecture.id} videoRef={videoRef}>
        <a
          href={courseNotesUrl(courseId)}
          download
          className="text-[13px] font-medium text-muted hover:text-ink"
        >
          Download course notes (.md)
        </a>
      </NoteComposer>
      {notes.isError && (
        <p className="text-sm text-danger" role="alert">
          Could not load notes ({notes.error.message}).
        </p>
      )}
      {notes.data !== undefined && notes.data.length > 0 && (
        <ol className="grid list-none">
          {notes.data.map((note) => (
            <NoteItem key={note.id} note={note} lectureId={lecture.id} onSeek={seekTo} />
          ))}
        </ol>
      )}
    </section>
  )
}
