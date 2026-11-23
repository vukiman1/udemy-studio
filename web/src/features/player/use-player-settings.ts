import { useEffect, useState } from 'react'
import { parsePlayerSettings, type PlayerSettings } from './player-settings'

const PLAYER_SETTINGS_STORAGE_KEY = 'udemy-studio.player-settings'

type SettingsUpdate = (current: PlayerSettings) => PlayerSettings

export function usePlayerSettings(): readonly [PlayerSettings, (update: SettingsUpdate) => void] {
  const [settings, setSettings] = useState<PlayerSettings>(() =>
    parsePlayerSettings(window.localStorage.getItem(PLAYER_SETTINGS_STORAGE_KEY)),
  )

  useEffect(() => {
    window.localStorage.setItem(PLAYER_SETTINGS_STORAGE_KEY, JSON.stringify(settings))
  }, [settings])

  return [settings, setSettings] as const
}
