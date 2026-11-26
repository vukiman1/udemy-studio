import { useMutation } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { CircleCheck, Eye, LoaderCircle } from 'lucide-react'
import { checkAI } from '../../api/ai-queries'
import { Button, ButtonVariant } from '../../ui/Button'
import { TEXT_FIELD } from '../../ui/text-field'
import { SETTINGS_LABEL, SETTINGS_ROW } from '../../ui/settings-row'
import type { AIDraftEditor } from './use-ai-draft'
import { DISPLAY_LOCALE } from '../../app/display-locale'

const MS_PER_SECOND = 1000
const SECONDS_FORMAT = new Intl.NumberFormat(DISPLAY_LOCALE, { maximumFractionDigits: 1 })
const CLOCK_FORMAT = new Intl.DateTimeFormat(DISPLAY_LOCALE, { timeStyle: 'short' })
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
          API key
        </span>
        <div className="flex flex-wrap items-center gap-2.5">
          {isEditingKey ? (
            <input
              className={`${TEXT_FIELD} min-w-0 flex-1 basis-60`}
              type="password"
              autoComplete="off"
              aria-labelledby="api-key-label"
              placeholder={`Paste a ${provider.label} key`}
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
              Cancel
            </Button>
          )}
          {!isEditingKey && (
            <Button icon={Eye} onClick={editor.startEditingKey}>
              Change key
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
            {check.isPending ? 'Checking…' : 'Check'}
          </Button>
        </div>
      </div>
      {check.isSuccess && (
        <p className="text-[13px] text-muted">
          <span className="text-accent-soft">● Key works.</span>
          {` ${provider.model.label} answered in ${SECONDS_FORMAT.format(check.data.result.latencyMs / MS_PER_SECOND)} s at ${CLOCK_FORMAT.format(check.data.at)}.`}
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
