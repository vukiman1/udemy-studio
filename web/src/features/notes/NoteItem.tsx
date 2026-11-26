import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type ReactElement } from 'react'
import { deleteNote, editNote } from '../../api/note-queries'
import { QUERY_KEYS } from '../../api/queries'
import type { Note } from '../../api/types'
import { MS_PER_SECOND, formatTimestamp } from '../../app/timestamp'
import { Button, ButtonSize, ButtonVariant } from '../../ui/Button'
import { isSubmitShortcut } from './submit-shortcut'

interface NoteItemProps {
  readonly note: Note
  readonly lectureId: string
  // Undefined for lectures without video: their notes have no moment to jump to.
  readonly onSeek: ((seconds: number) => void) | undefined
}

const TEXT_ACTION = 'rounded-none bg-transparent p-0 text-[13px] font-medium'
const QUIET_ACTION = `${TEXT_ACTION} text-muted hover:text-ink`
const DANGER_ACTION = `${TEXT_ACTION} text-danger`

export function NoteItem({ note, lectureId, onSeek }: NoteItemProps): ReactElement {
  const queryClient = useQueryClient()
  const [editedBody, setEditedBody] = useState<string | undefined>(undefined)
  const [isConfirmingDelete, setIsConfirmingDelete] = useState(false)
  const refreshNotes = (): Promise<void> =>
    queryClient.invalidateQueries({ queryKey: QUERY_KEYS.notes(lectureId) })
  const save = useMutation({
    mutationFn: (body: string) => editNote(note.id, body),
    onSuccess: async () => {
      setEditedBody(undefined)
      await refreshNotes()
    },
  })
  const remove = useMutation({ mutationFn: () => deleteNote(note.id), onSuccess: refreshNotes })
  const error = save.error ?? remove.error
  const canSave = editedBody !== undefined && editedBody.trim() !== '' && !save.isPending
  const saveEdit = (): void => {
    if (canSave) save.mutate(editedBody)
  }
  const stopEditing = (): void => {
    setEditedBody(undefined)
    save.reset()
  }

  return (
    <li className="flex items-start gap-3 border-b border-line py-2.5 last:border-b-0 motion-safe:animate-rise">
      {onSeek !== undefined && (
        <button
          type="button"
          className="rounded-none bg-transparent p-0 text-[13px] leading-5 font-semibold text-accent-soft tabular-nums"
          onClick={() => {
            onSeek(note.atSeconds)
          }}
        >
          {formatTimestamp(note.atSeconds * MS_PER_SECOND)}
        </button>
      )}
      <div className="grid min-w-0 flex-1 gap-2 text-sm leading-5">
        {editedBody === undefined ? (
          <p className="wrap-anywhere whitespace-pre-wrap">{note.body}</p>
        ) : (
          <form
            className="grid gap-2"
            onSubmit={(event) => {
              event.preventDefault()
              saveEdit()
            }}
          >
            <textarea
              className="min-h-20 rounded-xl px-3 py-2.5"
              autoFocus
              value={editedBody}
              onFocus={(event) => {
                const end = event.currentTarget.value.length
                event.currentTarget.setSelectionRange(end, end)
              }}
              onChange={(event) => {
                setEditedBody(event.target.value)
              }}
              onKeyDown={(event) => {
                if (!isSubmitShortcut(event)) return
                event.preventDefault()
                saveEdit()
              }}
            />
            <div className="flex gap-2">
              <Button
                type="submit"
                variant={ButtonVariant.Primary}
                size={ButtonSize.Small}
                disabled={!canSave}
              >
                Save
              </Button>
              <Button variant={ButtonVariant.Ghost} size={ButtonSize.Small} onClick={stopEditing}>
                Cancel
              </Button>
            </div>
          </form>
        )}
        {error !== null && (
          <p className="text-[13px] text-danger" role="alert">
            {error.message}
          </p>
        )}
      </div>
      <span className="flex shrink-0 gap-3.5 leading-5">
        {editedBody === undefined && (
          <button
            type="button"
            className={QUIET_ACTION}
            onClick={() => {
              setEditedBody(note.body)
            }}
          >
            Edit
          </button>
        )}
        {isConfirmingDelete ? (
          <>
            <button
              type="button"
              className={DANGER_ACTION}
              disabled={remove.isPending}
              onClick={() => {
                remove.mutate()
              }}
            >
              Delete
            </button>
            <button
              type="button"
              className={QUIET_ACTION}
              onClick={() => {
                setIsConfirmingDelete(false)
                remove.reset()
              }}
            >
              Keep
            </button>
          </>
        ) : (
          <button
            type="button"
            className={QUIET_ACTION}
            onClick={() => {
              setIsConfirmingDelete(true)
            }}
          >
            Delete
          </button>
        )}
      </span>
    </li>
  )
}
