import type { ReactElement } from 'react'
import type { AIProvider, AISettings } from '../../api/types'
import { ApiKeyField } from './ApiKeyField'
import { ModelPicker } from './ModelPicker'
import { SettingsCard } from './SettingsCard'
import { SettingsFooter } from './SettingsFooter'
import type { AIDraftEditor } from './use-ai-draft'

interface AIConfigCardProps {
  readonly providers: readonly AIProvider[]
  readonly saved: AISettings
  readonly editor: AIDraftEditor
}

export function AIConfigCard({ providers, saved, editor }: AIConfigCardProps): ReactElement {
  return (
    <SettingsCard title="Cấu hình">
      <label className="grid gap-2">
        <span className="font-semibold">Loại AI</span>
        <select
          className="rounded-lg border border-line bg-page px-3 py-2.5"
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
      <ApiKeyField editor={editor} />
      <ModelPicker editor={editor} />
      {/* Keyed by provider so a delete confirmation never carries over to another provider's key. */}
      <SettingsFooter key={editor.draft.provider} settings={saved} editor={editor} />
    </SettingsCard>
  )
}
