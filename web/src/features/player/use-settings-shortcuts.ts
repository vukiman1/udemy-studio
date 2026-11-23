import { useEffect, useEffectEvent } from 'react'

export interface SettingsShortcuts {
  readonly cycleSubtitles: () => void
  readonly stepRate: (direction: 1 | -1) => void
  readonly stepFontScale: (direction: 1 | -1) => void
}

const TYPING_ELEMENTS = new Set(['INPUT', 'SELECT', 'TEXTAREA'])

function bindings(shortcuts: SettingsShortcuts): Record<string, () => void> {
  return {
    c: shortcuts.cycleSubtitles,
    '<': () => {
      shortcuts.stepRate(-1)
    },
    '>': () => {
      shortcuts.stepRate(1)
    },
    '-': () => {
      shortcuts.stepFontScale(-1)
    },
    '=': () => {
      shortcuts.stepFontScale(1)
    },
  }
}

// Playback, seeking and fullscreen keys belong to the Video.js <Hotkey> elements; these cover our own settings.
export function useSettingsShortcuts(shortcuts: SettingsShortcuts): void {
  const onKeyDown = useEffectEvent((event: KeyboardEvent): void => {
    if (event.metaKey || event.ctrlKey || event.altKey || isTyping(event.target)) return
    const action = bindings(shortcuts)[event.key.toLowerCase()]
    if (action === undefined) return
    event.preventDefault()
    action()
  })

  useEffect(() => {
    window.addEventListener('keydown', onKeyDown)
    return () => {
      window.removeEventListener('keydown', onKeyDown)
    }
  }, [])
}

function isTyping(target: EventTarget | null): boolean {
  return (
    target instanceof HTMLElement &&
    (TYPING_ELEMENTS.has(target.tagName) || target.isContentEditable)
  )
}
