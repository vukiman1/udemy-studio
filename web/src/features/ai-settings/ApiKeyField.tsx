import { useMutation } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { CircleCheck, Eye, LoaderCircle } from 'lucide-react'
import { checkAI } from '../../api/ai-queries'
import { Button, ButtonVariant } from '../../ui/Button'
import { TEXT_FIELD } from '../../ui/text-field'
import { SETTINGS_LABEL, SETTINGS_ROW } from './settings-row'
import type { AIDraftEditor } from './use-ai-draft'

const MS_PER_SECOND = 1000
const SECONDS_FORMAT = new Intl.NumberFormat('vi-VN', { maximumFractionDigits: 1 })
const CLOCK_FORMAT = new Intl.DateTimeFormat('vi-VN', { timeStyle: 'short' })
const CHECKING_ICON_CLASS = '[&_svg]:animate-spin'

interface ApiKeyFieldProps {
  readonly editor: AIDraftEditor
}

export function ApiKeyField({ editor }: ApiKeyFieldProps): ReactElement {
  const { draft, provider, isEditingKey } = editor
  const check = useMutation({
    mutationFn: async () => ({
      result: await checkAI(draft.provider, draft.apiKey),
      at: new Date(),
    }),
  })
  const canCheck = !check.isPending && (!isEditingKey || draft.apiKey.trim() !== '')

  return (
    <div className="grid gap-3">
      <div className={SETTINGS_ROW}>
        <span className={SETTINGS_LABEL} id="api-key-label">
          Khoá API
        </span>
        <div className="flex flex-wrap items-center gap-2.5">
          {isEditingKey ? (
            <input
              className={`${TEXT_FIELD} min-w-0 flex-1 basis-60`}
              type="password"
              autoComplete="off"
              aria-labelledby="api-key-label"
              placeholder={`Dán khoá ${provider.label}`}
              value={draft.apiKey}
              onChange={(event) => {
                editor.typeKey(event.target.value)
                check.reset()
              }}
            />
          ) : (
            <span className="flex h-12 min-w-0 flex-1 basis-60 items-center truncate rounded-xl border border-line bg-input px-4 font-mono text-[13px] text-muted">
              {editor.saved?.keyHint}
            </span>
          )}
          {isEditingKey && editor.canKeepSavedKey && (
            <Button
              variant={ButtonVariant.Ghost}
              onClick={() => {
                editor.keepSavedKey()
                check.reset()
              }}
            >
              Huỷ
            </Button>
          )}
          {!isEditingKey && (
            <Button icon={Eye} onClick={editor.startEditingKey}>
              Đổi khoá
            </Button>
          )}
          <Button
            icon={check.isPending ? LoaderCircle : CircleCheck}
            className={check.isPending ? CHECKING_ICON_CLASS : ''}
            disabled={!canCheck}
            onClick={() => {
              check.mutate()
            }}
          >
            {check.isPending ? 'Đang kiểm tra…' : 'Kiểm tra'}
          </Button>
        </div>
      </div>
      {check.isSuccess && (
        <p className="text-[13px] text-muted">
          <span className="text-accent-soft">● Khoá dùng được.</span>
          {` ${provider.model.label} trả lời sau ${SECONDS_FORMAT.format(check.data.result.latencyMs / MS_PER_SECOND)} giây, lúc ${CLOCK_FORMAT.format(check.data.at)}.`}
        </p>
      )}
      {check.isError && (
        <p className="text-[13px] text-danger" role="alert">
          {check.error.message}
        </p>
      )}
    </div>
  )
}
