import type { ReactElement } from 'react'
import { Eye } from 'lucide-react'
import { Button, ButtonVariant } from '../../ui/Button'
import { SETTINGS_LABEL, SETTINGS_ROW } from '../../ui/settings-row'
import { TEXT_FIELD } from '../../ui/text-field'
import type { StorageDraftEditor } from './use-storage-draft'

const KEY_FIELD = `${TEXT_FIELD} min-w-0 flex-1 basis-60 font-mono`

interface StorageKeysFieldProps {
  readonly editor: StorageDraftEditor
}

export function StorageKeysField({ editor }: StorageKeysFieldProps): ReactElement {
  if (!editor.isEditingKeys)
    return (
      <div className={SETTINGS_ROW}>
        <span className={SETTINGS_LABEL}>Khoá truy cập</span>
        <div className="flex flex-wrap items-center gap-2.5">
          <span className="flex h-12 min-w-0 flex-1 basis-60 items-center truncate rounded-xl border border-line bg-input px-4 font-mono text-[13px] text-muted">
            {editor.saved.accessKeyIdHint}
          </span>
          <Button icon={Eye} onClick={editor.startEditingKeys}>
            Đổi khoá
          </Button>
        </div>
      </div>
    )

  return (
    <>
      <label className={SETTINGS_ROW}>
        <span className={SETTINGS_LABEL}>Access Key ID</span>
        <input
          className={KEY_FIELD}
          type="text"
          autoComplete="off"
          spellCheck={false}
          placeholder="Dán Access Key ID"
          value={editor.draft.accessKeyId}
          onChange={(event) => {
            editor.typeAccessKeyId(event.target.value)
          }}
        />
      </label>
      <div className={SETTINGS_ROW}>
        <span className={SETTINGS_LABEL} id="secret-access-key-label">
          Secret Access Key
        </span>
        <div className="flex flex-wrap items-center gap-2.5">
          <input
            className={KEY_FIELD}
            type="password"
            autoComplete="off"
            aria-labelledby="secret-access-key-label"
            placeholder="Dán Secret Access Key"
            value={editor.draft.secretAccessKey}
            onChange={(event) => {
              editor.typeSecretAccessKey(event.target.value)
            }}
          />
          {editor.canKeepSavedKeys && (
            <Button variant={ButtonVariant.Ghost} onClick={editor.keepSavedKeys}>
              Huỷ
            </Button>
          )}
        </div>
      </div>
    </>
  )
}
