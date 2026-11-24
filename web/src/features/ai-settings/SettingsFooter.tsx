import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type ReactElement } from 'react'
import { forgetAIProvider, saveAISettings, setAIEnabled } from '../../api/ai-queries'
import { QUERY_KEYS } from '../../api/queries'
import type { AISettings } from '../../api/types'
import type { AIDraftEditor } from './use-ai-draft'

interface SettingsFooterProps {
  readonly settings: AISettings
  readonly editor: AIDraftEditor
}

export function SettingsFooter({ settings, editor }: SettingsFooterProps): ReactElement {
  const queryClient = useQueryClient()
  const [isConfirmingDelete, setIsConfirmingDelete] = useState(false)
  const keep = (saved: AISettings): void => {
    queryClient.setQueryData(QUERY_KEYS.aiSettings, saved)
  }
  const save = useMutation({ mutationFn: () => saveAISettings(editor.draft), onSuccess: keep })
  const toggle = useMutation({ mutationFn: () => setAIEnabled(!settings.enabled), onSuccess: keep })
  const forget = useMutation({
    mutationFn: () => forgetAIProvider(editor.draft.provider),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: QUERY_KEYS.aiSettings }),
  })
  const error = save.error ?? toggle.error ?? forget.error
  const canSave =
    editor.isDirty && (!editor.isEditingKey || editor.draft.apiKey.trim() !== '') && !save.isPending

  return (
    <div className="grid gap-2 border-t border-line pt-5">
      <div className="flex flex-wrap items-center gap-4">
        <button
          type="button"
          disabled={!canSave}
          onClick={() => {
            save.mutate()
          }}
        >
          {save.isPending ? 'Đang kiểm tra & lưu…' : 'Lưu'}
        </button>
        <span className="flex-1 text-[0.9rem] text-muted">{statusText(settings, editor)}</span>
        {editor.isInUse && (
          <button
            type="button"
            disabled={toggle.isPending}
            onClick={() => {
              toggle.mutate()
            }}
          >
            {settings.enabled ? 'Ngừng dùng' : 'Dùng lại'}
          </button>
        )}
        {editor.saved !== undefined &&
          (isConfirmingDelete ? (
            <button
              type="button"
              className="text-danger"
              disabled={forget.isPending}
              onClick={() => {
                forget.mutate()
              }}
            >
              Xoá hẳn khoá?
            </button>
          ) : (
            <button
              type="button"
              className="text-danger"
              onClick={() => {
                setIsConfirmingDelete(true)
              }}
            >
              Xoá khoá
            </button>
          ))}
      </div>
      {error !== null && (
        <p className="m-0 text-[0.9rem] text-danger" role="alert">
          {error.message}
        </p>
      )}
    </div>
  )
}

function statusText(settings: AISettings, editor: AIDraftEditor): string {
  const label = editor.provider.label
  const hasOnlySwitched =
    !editor.isInUse && editor.saved !== undefined && editor.draft.apiKey.trim() === ''
  if (editor.isInUse && !editor.isDirty)
    return settings.enabled
      ? `Trang web đang dùng ${label} với cấu hình này.`
      : `Đang ngừng dùng ${label}.`
  if (hasOnlySwitched) return `Đã lưu khoá ${label}; bấm Lưu để chuyển sang dùng ${label}.`
  if (!settings.configured && editor.saved === undefined) return 'Chưa cấu hình AI nào.'
  return 'Có thay đổi chưa lưu.'
}
