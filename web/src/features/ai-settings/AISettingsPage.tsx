import { ChevronIcon } from '@videojs/react/icons'
import type { ReactElement } from 'react'
import { Link } from 'react-router'
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
    <main className="mx-auto grid max-w-[760px] gap-4 px-4 py-6">
      <header className="flex items-center gap-3">
        <Link to="/" className="inline-flex items-center gap-1 text-muted">
          <ChevronIcon className="size-4.5 -scale-x-100" />
          Thư viện
        </Link>
        <h1 className="text-2xl font-bold">Cài đặt AI</h1>
      </header>
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
