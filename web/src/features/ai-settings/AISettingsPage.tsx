import type { ReactElement } from 'react'
import { useAIProviders, useAISettings } from '../../api/ai-queries'
import type { AIProvider, AISettings } from '../../api/types'
import { PageStatus } from '../../app/PageStatus'
import { AIConfigCard } from './AIConfigCard'
import { useAIDraft } from './use-ai-draft'

export function AISettingsPage(): ReactElement {
  const providers = useAIProviders()
  const settings = useAISettings()
  const error = providers.error ?? settings.error

  return (
    <main className="grid max-w-[1160px] content-start gap-7 px-5 py-8 md:px-10 md:py-10">
      <h1 className="text-[30px] font-semibold">Cài đặt AI</h1>
      {error !== null && <PageStatus tone="error" message={error.message} />}
      {providers.data?.[0] !== undefined && settings.data !== undefined && (
        <AISettingsEditor
          key={settingsVersion(settings.data)}
          providers={providers.data}
          defaultProvider={providers.data[0]}
          saved={settings.data}
        />
      )}
    </main>
  )
}

interface AISettingsEditorProps {
  readonly providers: readonly AIProvider[]
  readonly defaultProvider: AIProvider
  readonly saved: AISettings
}

function AISettingsEditor({
  providers,
  defaultProvider,
  saved,
}: AISettingsEditorProps): ReactElement {
  const editor = useAIDraft(providers, defaultProvider, saved)
  return <AIConfigCard providers={providers} saved={saved} editor={editor} />
}

// A new version of the saved settings starts the form over from it.
function settingsVersion(settings: AISettings): string {
  const saved = settings.providers.map((provider) => `${provider.provider}:${provider.keyHint}`)
  return [settings.provider, String(settings.enabled), ...saved].join('|')
}
