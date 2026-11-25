export const PLAYBACK_RATES = [
  0.5, 0.6, 0.7, 0.75, 0.8, 0.85, 0.9, 0.95, 1, 1.1, 1.25, 1.5, 1.75, 2,
] as const
export const FONT_SCALES = [0.8, 1, 1.2, 1.4, 1.7] as const
export const TEXT_OPACITIES = [0.25, 0.5, 0.75, 1] as const
export const BACKGROUND_OPACITIES = [0, 0.25, 0.5, 0.75, 1] as const
export const PRIMARY_LANGUAGE = 'vi'
export const SECONDARY_LANGUAGE = 'en'

export enum SubtitleMode {
  Bilingual = 'bilingual',
  Primary = 'primary',
  Secondary = 'secondary',
  Off = 'off',
}

const SUBTITLE_MODE_CYCLE = [
  SubtitleMode.Bilingual,
  SubtitleMode.Primary,
  SubtitleMode.Secondary,
  SubtitleMode.Off,
] as const

export interface PlayerSettings {
  readonly playbackRate: number
  readonly subtitleMode: SubtitleMode
  readonly fontScale: number
  readonly textOpacity: number
  readonly backgroundOpacity: number
}

export const DEFAULT_PLAYER_SETTINGS: PlayerSettings = {
  playbackRate: 1,
  subtitleMode: SubtitleMode.Bilingual,
  fontScale: 1,
  textOpacity: 1,
  backgroundOpacity: 0.75,
}

export function parsePlayerSettings(stored: string | null): PlayerSettings {
  if (stored === null) return DEFAULT_PLAYER_SETTINGS
  let candidate: unknown
  try {
    candidate = JSON.parse(stored)
  } catch {
    return DEFAULT_PLAYER_SETTINGS
  }
  if (typeof candidate !== 'object' || candidate === null) return DEFAULT_PLAYER_SETTINGS
  const record = candidate as Record<string, unknown>
  return {
    playbackRate: isListed(PLAYBACK_RATES, record.playbackRate)
      ? record.playbackRate
      : DEFAULT_PLAYER_SETTINGS.playbackRate,
    subtitleMode: parseSubtitleMode(record.subtitleMode) ?? DEFAULT_PLAYER_SETTINGS.subtitleMode,
    fontScale: isListed(FONT_SCALES, record.fontScale)
      ? record.fontScale
      : DEFAULT_PLAYER_SETTINGS.fontScale,
    textOpacity: isListed(TEXT_OPACITIES, record.textOpacity)
      ? record.textOpacity
      : DEFAULT_PLAYER_SETTINGS.textOpacity,
    backgroundOpacity: isListed(BACKGROUND_OPACITIES, record.backgroundOpacity)
      ? record.backgroundOpacity
      : DEFAULT_PLAYER_SETTINGS.backgroundOpacity,
  }
}

export function languagesFor(mode: SubtitleMode, available: readonly string[]): string[] {
  const has = (language: string): boolean => available.includes(language)
  switch (mode) {
    case SubtitleMode.Bilingual:
      return [PRIMARY_LANGUAGE, SECONDARY_LANGUAGE].filter(has)
    case SubtitleMode.Primary:
      return [PRIMARY_LANGUAGE].filter(has)
    case SubtitleMode.Secondary:
      return [SECONDARY_LANGUAGE].filter(has)
    case SubtitleMode.Off:
      return []
  }
}

export function nextSubtitleMode(mode: SubtitleMode): SubtitleMode {
  const position = SUBTITLE_MODE_CYCLE.indexOf(mode)
  return SUBTITLE_MODE_CYCLE[(position + 1) % SUBTITLE_MODE_CYCLE.length] ?? SubtitleMode.Bilingual
}

export function toggleSubtitles(mode: SubtitleMode): SubtitleMode {
  return mode === SubtitleMode.Off ? SubtitleMode.Bilingual : SubtitleMode.Off
}

export function stepThrough(steps: readonly number[], current: number, direction: 1 | -1): number {
  const position = steps.indexOf(current)
  return steps[Math.min(Math.max(position + direction, 0), steps.length - 1)] ?? current
}

function isListed(steps: readonly number[], value: unknown): value is number {
  return typeof value === 'number' && steps.includes(value)
}

export function parseSubtitleMode(value: unknown): SubtitleMode | undefined {
  return Object.values(SubtitleMode).find((mode) => mode === value)
}
