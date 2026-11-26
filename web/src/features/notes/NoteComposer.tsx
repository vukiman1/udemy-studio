import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type ReactElement, type ReactNode, type RefObject } from 'react'
import { Save } from 'lucide-react'
import { addNote } from '../../api/note-queries'
import { QUERY_KEYS } from '../../api/queries'
import { MS_PER_SECOND, formatTimestamp } from '../../app/timestamp'
import { Button, ButtonSize, ButtonVariant } from '../../ui/Button'
import { isSubmitShortcut } from './submit-shortcut'

interface NoteComposerProps {
  readonly lectureId: string
  readonly videoRef: RefObject<HTMLVideoElement | null> | null
  readonly children?: ReactNode
}

const LECTURE_START_SECONDS = 0

export function NoteComposer({ lectureId, videoRef, children }: NoteComposerProps): ReactElement {
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
      className="grid gap-3"
      onSubmit={(event) => {
        event.preventDefault()
        if (canSave) save.mutate()
      }}
    >
      <div className="flex flex-wrap items-center gap-3">
        <h3 className="mr-auto text-[15px] font-semibold text-muted">Notes</h3>
        {children}
        <Button
          type="submit"
          variant={ButtonVariant.Primary}
          size={ButtonSize.Small}
          icon={Save}
          disabled={!canSave}
        >
          {videoRef === null || draft === ''
            ? 'Save note'
            : `Save note at ${formatTimestamp(atSeconds * MS_PER_SECOND)}`}
        </Button>
      </div>
      <textarea
        className="min-h-24 rounded-xl px-4 py-3.5 text-sm placeholder:text-muted"
        placeholder="Write a note… (⌘/Ctrl + Enter to save)"
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
      {save.isError && (
        <p className="text-sm text-danger" role="alert">
          {save.error.message}
        </p>
      )}
    </form>
  )
}
