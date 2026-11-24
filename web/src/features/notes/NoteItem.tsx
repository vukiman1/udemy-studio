import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type ReactElement } from 'react'
import { deleteNote, editNote } from '../../api/note-queries'
import { QUERY_KEYS } from '../../api/queries'
import type { Note } from '../../api/types'
import { MS_PER_SECOND, formatTimestamp } from '../../app/timestamp'
import { isSubmitShortcut } from './submit-shortcut'

interface NoteItemProps {
  readonly note: Note
  readonly lectureId: string
  // Undefined for lectures without video: their notes have no moment to jump to.
  readonly onSeek: ((seconds: number) => void) | undefined
}

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
    <li className="grid gap-1.5 rounded-lg border border-line bg-page p-3">
      <div className="flex flex-wrap items-center gap-3 text-[0.85rem]">
        {onSeek !== undefined && (
          <button
            type="button"
            className="font-semibold text-accent"
            onClick={() => {
              onSeek(note.atSeconds)
            }}
          >
            {formatTimestamp(note.atSeconds * MS_PER_SECOND)}
          </button>
        )}
        <span className="ml-auto flex gap-3">
          {editedBody === undefined && (
            <button
              type="button"
              onClick={() => {
                setEditedBody(note.body)
              }}
            >
              Sửa
            </button>
          )}
          {isConfirmingDelete ? (
            <>
              <button
                type="button"
                className="text-danger"
                disabled={remove.isPending}
                onClick={() => {
                  remove.mutate()
                }}
              >
                Xoá hẳn
              </button>
              <button
                type="button"
                onClick={() => {
                  setIsConfirmingDelete(false)
                  remove.reset()
                }}
              >
                Giữ lại
              </button>
            </>
          ) : (
            <button
              type="button"
              onClick={() => {
                setIsConfirmingDelete(true)
              }}
            >
              Xoá
            </button>
          )}
        </span>
      </div>
      {editedBody === undefined ? (
        <p className="m-0 whitespace-pre-wrap">{note.body}</p>
      ) : (
        <form
          className="grid gap-2"
          onSubmit={(event) => {
            event.preventDefault()
            saveEdit()
          }}
        >
          <textarea
            className="min-h-20 rounded-md border border-line bg-surface px-2.5 py-1.5"
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
          <div className="flex gap-3">
            <button type="submit" disabled={!canSave}>
              Lưu
            </button>
            <button type="button" onClick={stopEditing}>
              Huỷ
            </button>
          </div>
        </form>
      )}
      {error !== null && (
        <p className="m-0 text-[0.9rem] text-danger" role="alert">
          {error.message}
        </p>
      )}
    </li>
  )
}
