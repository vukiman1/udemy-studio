import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type ReactElement } from 'react'
import { Pause, Play, Save, Trash2 } from 'lucide-react'
import { forgetAIProvider, saveAISettings, setAIEnabled } from '../../api/ai-queries'
import { QUERY_KEYS } from '../../api/queries'
import type { AISettings } from '../../api/types'
import { Button, ButtonVariant } from '../../ui/Button'
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
    <div className="grid gap-2 border-t border-line-strong pt-5">
      <div className="flex flex-wrap items-center gap-2.5">
        <FooterStatus settings={settings} editor={editor} />
        <Button
          variant={canSave ? ButtonVariant.Primary : ButtonVariant.Secondary}
          icon={Save}
          disabled={!canSave}
          onClick={() => {
            save.mutate()
          }}
        >
          {save.isPending ? 'Đang kiểm tra & lưu…' : 'Lưu'}
        </Button>
        {editor.isInUse && (
          <Button
            variant={ButtonVariant.Ghost}
            icon={settings.enabled ? Pause : Play}
            disabled={toggle.isPending}
            onClick={() => {
              toggle.mutate()
            }}
          >
            {settings.enabled ? 'Dừng' : 'Dùng lại'}
          </Button>
        )}
        {editor.saved !== undefined && (
          <Button
            variant={ButtonVariant.Danger}
            icon={Trash2}
            disabled={forget.isPending}
            onClick={() => {
              if (isConfirmingDelete) forget.mutate()
              else setIsConfirmingDelete(true)
            }}
          >
            {isConfirmingDelete ? 'Xoá hẳn khoá?' : 'Xoá'}
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
  readonly settings: AISettings
  readonly editor: AIDraftEditor
}

function FooterStatus({ settings, editor }: FooterStatusProps): ReactElement {
  if (editor.isInUse && !editor.isDirty && settings.enabled)
    return (
      <span className="mr-auto flex items-center gap-1.5 text-xs font-semibold text-success">
        <span className="size-2 rounded-full bg-success" aria-hidden="true" />
        Đang sử dụng
      </span>
    )
  return <span className="mr-auto text-xs text-muted">{statusText(settings, editor)}</span>
}

function statusText(settings: AISettings, editor: AIDraftEditor): string {
  const label = editor.provider.label
  const hasOnlySwitched =
    !editor.isInUse && editor.saved !== undefined && editor.draft.apiKey.trim() === ''
  if (editor.isInUse && !editor.isDirty && !settings.enabled) return `Đang ngừng dùng ${label}.`
  if (hasOnlySwitched) return `Đã lưu khoá ${label}; bấm Lưu để chuyển sang dùng ${label}.`
  if (!settings.configured && editor.saved === undefined) return 'Chưa cấu hình AI nào.'
  return 'Có thay đổi chưa lưu.'
}
