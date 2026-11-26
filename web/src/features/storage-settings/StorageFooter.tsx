import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type ReactElement } from 'react'
import { Save, Trash2 } from 'lucide-react'
import { QUERY_KEYS } from '../../api/queries'
import { forgetStorageSettings, saveStorageSettings } from '../../api/storage-queries'
import type { StorageSettings } from '../../api/types'
import { Button, ButtonVariant } from '../../ui/Button'
import type { StorageDraftEditor } from './use-storage-draft'

interface StorageFooterProps {
  readonly editor: StorageDraftEditor
}

export function StorageFooter({ editor }: StorageFooterProps): ReactElement {
  const queryClient = useQueryClient()
  const [isConfirmingDelete, setIsConfirmingDelete] = useState(false)
  const save = useMutation({
    mutationFn: saveStorageSettings,
    onSuccess: (saved: StorageSettings) => {
      queryClient.setQueryData(QUERY_KEYS.storageSettings, saved)
      void queryClient.invalidateQueries({ queryKey: QUERY_KEYS.storageUsage })
    },
  })
  const forget = useMutation({
    mutationFn: forgetStorageSettings,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: QUERY_KEYS.storageSettings }),
  })
  const error = save.error ?? forget.error
  const canSave = editor.isDirty && editor.isComplete && !save.isPending

  return (
    <div className="grid gap-2 border-t border-line-strong pt-5">
      <div className="flex flex-wrap items-center gap-2.5">
        <FooterStatus editor={editor} />
        <Button
          variant={canSave ? ButtonVariant.Primary : ButtonVariant.Secondary}
          icon={Save}
          disabled={!canSave}
          onClick={() => {
            save.mutate(editor.draft)
          }}
        >
          {save.isPending ? 'Checking & saving…' : 'Save'}
        </Button>
        {editor.saved.configured && (
          <Button
            variant={ButtonVariant.Danger}
            icon={Trash2}
            disabled={forget.isPending}
            onClick={() => {
              if (isConfirmingDelete) forget.mutate()
              else setIsConfirmingDelete(true)
            }}
          >
            {isConfirmingDelete ? 'Delete connection?' : 'Delete'}
          </Button>
        )}
      </div>
      {error !== null && (
        <p className="text-[13px] text-danger" role="alert">
          {error.message}
        </p>
      )}
    </div>
  )
}

interface FooterStatusProps {
  readonly editor: StorageDraftEditor
}

function FooterStatus({ editor }: FooterStatusProps): ReactElement {
  if (editor.saved.configured && !editor.isDirty)
    return (
      <span className="mr-auto flex items-center gap-1.5 text-xs font-semibold text-success">
        <span className="size-2 rounded-full bg-success" aria-hidden="true" />
        Connected to {bucketName(editor.saved.url)}
      </span>
    )
  return (
    <span className="mr-auto text-xs text-muted">
      {editor.saved.configured || editor.isComplete ? 'Unsaved changes.' : 'No bucket connected.'}
    </span>
  )
}

function bucketName(url: string | undefined): string {
  return url?.split('/').at(-1) ?? ''
}
