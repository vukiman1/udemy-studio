import type { ReactElement } from 'react'
import { useStorageSettings } from '../../api/storage-queries'
import type { StorageSettings } from '../../api/types'
import { PageStatus } from '../../app/PageStatus'
import { SettingsCard } from '../../ui/SettingsCard'
import { SETTINGS_LABEL, SETTINGS_ROW } from '../../ui/settings-row'
import { TEXT_FIELD } from '../../ui/text-field'
import { StorageCheckField } from './StorageCheckField'
import { StorageFooter } from './StorageFooter'
import { StorageKeysField } from './StorageKeysField'
import { useStorageDraft } from './use-storage-draft'

const URL_PLACEHOLDER = 'https://<account-id>.r2.cloudflarestorage.com/<bucket>'

export function StorageSettingsPage(): ReactElement {
  const settings = useStorageSettings()

  return (
    <main className="grid max-w-[1160px] content-start gap-7 px-5 py-8 md:px-10 md:py-10">
      <div className="grid gap-2">
        <h1 className="text-[30px] font-semibold">Lưu trữ đám mây</h1>
        <p className="text-sm text-muted">
          Kết nối một bucket Cloudflare R2 (hoặc S3) để đưa khoá đang học lên và học ở mọi nơi.
        </p>
      </div>
      {settings.isError && <PageStatus tone="error" message={settings.error.message} />}
      {settings.data !== undefined && (
        <StorageEditor key={settingsVersion(settings.data)} saved={settings.data} />
      )}
    </main>
  )
}

interface StorageEditorProps {
  readonly saved: StorageSettings
}

function StorageEditor({ saved }: StorageEditorProps): ReactElement {
  const editor = useStorageDraft(saved)
  return (
    <SettingsCard title="Kết nối">
      <label className={SETTINGS_ROW}>
        <span className={SETTINGS_LABEL}>S3 API URL</span>
        <input
          className={`${TEXT_FIELD} font-mono`}
          type="url"
          autoComplete="off"
          spellCheck={false}
          placeholder={URL_PLACEHOLDER}
          value={editor.draft.url}
          onChange={(event) => {
            editor.typeUrl(event.target.value)
          }}
        />
      </label>
      <StorageKeysField editor={editor} />
      <StorageCheckField editor={editor} />
      <StorageFooter editor={editor} />
    </SettingsCard>
  )
}

function settingsVersion(settings: StorageSettings): string {
  return [String(settings.configured), settings.url, settings.accessKeyIdHint].join('|')
}
