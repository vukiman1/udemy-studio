import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type ReactElement, type RefObject } from 'react'
import { addNote } from '../../api/note-queries'
import { QUERY_KEYS } from '../../api/queries'
import { MS_PER_SECOND, formatTimestamp } from '../../app/timestamp'
import { isSubmitShortcut } from './submit-shortcut'

interface NoteComposerProps {
  readonly lectureId: string
  readonly videoRef: RefObject<HTMLVideoElement | null> | null
}

const LECTURE_START_SECONDS = 0

export function NoteComposer({ lectureId, videoRef }: NoteComposerProps): ReactElement {
  const queryClient = useQueryClient()
  const [draft, setDraft] = useState('')
  // The moment is taken at the first keystroke, not at saving, since the video keeps playing meanwhile.
  const [atSeconds, setAtSeconds] = useState(LECTURE_START_SECONDS)
  const save = useMutation({
    mutationFn: () => addNote(lectureId, atSeconds, draft),
    onSuccess: async () => {
      setDraft('')
      await queryClient.invalidateQueries({ queryKey: QUERY_KEYS.notes(lectureId) })
    },
  })
  const canSave = draft.trim() !== '' && !save.isPending

  return (
    <form
      className="grid gap-2"
      onSubmit={(event) => {
        event.preventDefault()
        if (canSave) save.mutate()
      }}
    >
      <textarea
        className="min-h-20 rounded-md border border-line bg-page px-2.5 py-1.5"
        placeholder="Viết ghi chú… (⌘/Ctrl + Enter để lưu)"
        value={draft}
        readOnly={save.isPending}
        onChange={(event) => {
          if (draft === '') setAtSeconds(videoRef?.current?.currentTime ?? LECTURE_START_SECONDS)
          setDraft(event.target.value)
        }}
        onKeyDown={(event) => {
          if (!isSubmitShortcut(event) || !canSave) return
          event.preventDefault()
          save.mutate()
        }}
      />
      <div className="flex items-center gap-3">
        <button type="submit" disabled={!canSave}>
          {videoRef === null || draft === ''
            ? 'Lưu ghi chú'
            : `Lưu ghi chú tại ${formatTimestamp(atSeconds * MS_PER_SECOND)}`}
        </button>
        {save.isError && <p className="m-0 text-danger">{save.error.message}</p>}
      </div>
    </form>
  )
}
