import { describe, expect, it } from 'vitest'
import {
  DEFAULT_PLAYER_SETTINGS,
  FONT_SCALES,
  PLAYBACK_RATES,
  SubtitleMode,
  languagesFor,
  nextSubtitleMode,
  parsePlayerSettings,
  stepThrough,
  toggleSubtitles,
} from './player-settings'

describe('parsePlayerSettings', () => {
  it('reads back stored settings', () => {
    const stored = JSON.stringify({
      playbackRate: 1.5,
      subtitleMode: 'primary',
      fontScale: 1.4,
      textOpacity: 0.75,
      backgroundOpacity: 0,
    })
    expect(parsePlayerSettings(stored)).toEqual({
      playbackRate: 1.5,
      subtitleMode: SubtitleMode.Primary,
      fontScale: 1.4,
      textOpacity: 0.75,
      backgroundOpacity: 0,
    })
  })

  it.each([null, 'not json', '42', 'null'])('falls back to defaults for %s', (stored) => {
    expect(parsePlayerSettings(stored)).toEqual(DEFAULT_PLAYER_SETTINGS)
  })

  it('replaces each invalid field on its own', () => {
    const stored = JSON.stringify({
      playbackRate: 3,
      subtitleMode: 'karaoke',
      fontScale: 99,
      textOpacity: 0,
      backgroundOpacity: 2,
    })
    expect(parsePlayerSettings(stored)).toEqual(DEFAULT_PLAYER_SETTINGS)
  })
})

describe('languagesFor', () => {
  it('stacks Vietnamese above English', () => {
    expect(languagesFor(SubtitleMode.Bilingual, ['en', 'vi'])).toEqual(['vi', 'en'])
  })
  it('keeps what exists when a language is missing', () => {
    expect(languagesFor(SubtitleMode.Bilingual, ['en'])).toEqual(['en'])
  })
  it('shows nothing when off', () => {
    expect(languagesFor(SubtitleMode.Off, ['en', 'vi'])).toEqual([])
  })
})

describe('mode and rate stepping', () => {
  it('cycles through every subtitle mode', () => {
    expect(nextSubtitleMode(SubtitleMode.Bilingual)).toBe(SubtitleMode.Primary)
    expect(nextSubtitleMode(SubtitleMode.Off)).toBe(SubtitleMode.Bilingual)
  })
  it('steps through a list and stops at the ends', () => {
    expect(stepThrough(PLAYBACK_RATES, 1, 1)).toBe(1.25)
    expect(stepThrough(PLAYBACK_RATES, 2, 1)).toBe(2)
    expect(stepThrough(FONT_SCALES, 0.8, -1)).toBe(0.8)
  })
})

describe('toggleSubtitles', () => {
  it('turns any visible mode off', () => {
    expect(toggleSubtitles(SubtitleMode.Primary)).toBe(SubtitleMode.Off)
  })
  it('turns subtitles back on as bilingual', () => {
    expect(toggleSubtitles(SubtitleMode.Off)).toBe(SubtitleMode.Bilingual)
  })
})
