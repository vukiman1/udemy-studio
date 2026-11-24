import { useMutation } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { checkAI } from '../../api/ai-queries'
import type { AIDraftEditor } from './use-ai-draft'

const MS_PER_SECOND = 1000
const SECONDS_FORMAT = new Intl.NumberFormat('vi-VN', { maximumFractionDigits: 1 })
const CLOCK_FORMAT = new Intl.DateTimeFormat('vi-VN', { timeStyle: 'medium' })

interface ApiKeyFieldProps {
  readonly editor: AIDraftEditor
}

export function ApiKeyField({ editor }: ApiKeyFieldProps): ReactElement {
  const { draft, provider, isEditingKey } = editor
  const firstModel = draft.models[0]
  const check = useMutation({
    mutationFn: async () => ({
      result: await checkAI(draft.provider, draft.apiKey, firstModel),
      at: new Date(),
    }),
  })
  const labelOf = (id: string): string =>
    provider.models.find((model) => model.id === id)?.label ?? id
  const canCheck = !check.isPending && (!isEditingKey || draft.apiKey.trim() !== '')

  return (
    <div className="grid gap-2">
      <span className="font-semibold">Khoá API</span>
      <div className="flex rounded-lg border border-line bg-page">
        {isEditingKey ? (
          <span className="flex min-w-0 flex-1 items-center gap-3 pr-3">
            <input
              className="min-w-0 flex-1 rounded-l-lg bg-transparent px-3 py-2.5"
              type="password"
              autoComplete="off"
              aria-label={`Khoá API ${provider.label}`}
              placeholder={`Dán khoá ${provider.label}`}
              value={draft.apiKey}
              onChange={(event) => {
                editor.typeKey(event.target.value)
                check.reset()
              }}
            />
            {editor.canKeepSavedKey && (
              <button
                type="button"
                className="link-button"
                onClick={() => {
                  editor.keepSavedKey()
                  check.reset()
                }}
              >
                Huỷ
              </button>
            )}
          </span>
        ) : (
          <span className="flex min-w-0 flex-1 items-center justify-between gap-3 px-3 py-2.5">
            <span className="truncate font-mono">{editor.saved?.keyHint}</span>
            <button type="button" className="link-button" onClick={editor.startEditingKey}>
              Đổi khoá
            </button>
          </span>
        )}
        <button
          type="button"
          className="border-l border-line px-4"
          disabled={!canCheck}
          onClick={() => {
            check.mutate()
          }}
        >
          {check.isPending ? 'Đang kiểm tra…' : 'Kiểm tra'}
        </button>
      </div>
      {check.isSuccess && (
        <p className="m-0 text-[0.9rem]">
          <span className="text-accent">● Khoá dùng được.</span>
          {check.data.result.model !== undefined &&
            ` ${labelOf(check.data.result.model)} trả lời sau ${SECONDS_FORMAT.format((check.data.result.latencyMs ?? 0) / MS_PER_SECOND)} giây, lúc ${CLOCK_FORMAT.format(check.data.at)}.`}
        </p>
      )}
      {check.isError && (
        <p className="m-0 text-[0.9rem] text-danger" role="alert">
          {check.error.message}
        </p>
      )}
    </div>
  )
}
