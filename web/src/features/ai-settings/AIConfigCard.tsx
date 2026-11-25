import type { ReactElement } from 'react'
import type { AIProvider, AISettings } from '../../api/types'
import { ApiKeyField } from './ApiKeyField'
import { ProviderModel } from './ProviderModel'
import { SettingsCard } from './SettingsCard'
import { SettingsFooter } from './SettingsFooter'
import { SETTINGS_LABEL, SETTINGS_ROW } from './settings-row'
import type { AIDraftEditor } from './use-ai-draft'

interface AIConfigCardProps {
  readonly providers: readonly AIProvider[]
  readonly saved: AISettings
  readonly editor: AIDraftEditor
}

export function AIConfigCard({ providers, saved, editor }: AIConfigCardProps): ReactElement {
  return (
    <SettingsCard title="Cấu hình">
      <label className={SETTINGS_ROW}>
        <span className={SETTINGS_LABEL}>Loại AI</span>
        <select
          className="h-[50px] rounded-[11px] border-line-strong pr-4 pl-[18px] text-sm font-semibold"
          value={editor.draft.provider}
          onChange={(event) => {
            const chosen = providers.find((provider) => provider.id === event.target.value)
            if (chosen !== undefined) editor.chooseProvider(chosen.id)
          }}
        >
          {providers.map((provider) => (
            <option key={provider.id} value={provider.id}>
              {provider.label}
            </option>
          ))}
        </select>
      </label>
      {/* Keyed by provider so a check result or delete confirmation never carries over to another provider. */}
      <ApiKeyField key={`key-${editor.draft.provider}`} editor={editor} />
      <ProviderModel provider={editor.provider} />
      <SettingsFooter key={`footer-${editor.draft.provider}`} settings={saved} editor={editor} />
    </SettingsCard>
  )
}
