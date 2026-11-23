import { describe, expect, it } from 'vitest'
import type { Cue } from '../../api/types'
import { activeLinesAt } from './active-cues-at'

function cue(startMs: number, endMs: number, language: string, text = language): Cue {
  return { startMs, endMs, lines: [{ language, text }] }
}

const ORDER = ['vi', 'en']

describe('activeLinesAt', () => {
  it('shows both languages when their timings differ by a few milliseconds', () => {
    const cues = [cue(1000, 3000, 'vi'), cue(1000, 3010, 'en')]
    expect(activeLinesAt(cues, 2000, ORDER).map((line) => line.language)).toEqual(['vi', 'en'])
  })

  it('keeps a long cue visible after a shorter later cue ended', () => {
    const cues = [cue(0, 5000, 'en'), cue(1000, 3000, 'vi')]
    expect(activeLinesAt(cues, 4000, ORDER).map((line) => line.language)).toEqual(['en'])
  })

  it('orders lines by the requested languages', () => {
    const cues = [cue(0, 2000, 'en'), cue(0, 2000, 'vi')]
    expect(activeLinesAt(cues, 500, ORDER).map((line) => line.language)).toEqual(['vi', 'en'])
  })

  it('shows merged bilingual cues as they are', () => {
    const merged: Cue = {
      startMs: 0,
      endMs: 1000,
      lines: [
        { language: 'vi', text: 'Xin chào' },
        { language: 'en', text: 'Hello' },
      ],
    }
    expect(activeLinesAt([merged], 999, ORDER)).toEqual(merged.lines)
  })

  it.each([-1, 1000, 2500, 99999])('shows nothing at %ims outside every cue', (timeMs) => {
    expect(activeLinesAt([cue(0, 1000, 'en'), cue(2000, 2400, 'en')], timeMs, ORDER)).toEqual([])
  })

  it('handles an empty list', () => {
    expect(activeLinesAt([], 10, ORDER)).toEqual([])
  })
})
