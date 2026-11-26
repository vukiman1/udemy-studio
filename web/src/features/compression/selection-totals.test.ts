import { describe, expect, it } from 'vitest'
import type { CompressionChapter } from '../../api/types'
import { selectionTotals } from './selection-totals'

const CHAPTERS: readonly CompressionChapter[] = [
  { dir: '01 - Intro', videos: 2, totalBytes: 1_000, estimatedSeconds: 60 },
  { dir: '02 - Deep', videos: 3, totalBytes: 5_000, estimatedSeconds: 600 },
]

describe('selectionTotals', () => {
  it('adds up only the chosen chapters', () => {
    expect(selectionTotals(CHAPTERS, new Set(['02 - Deep']))).toEqual({
      videos: 3,
      totalBytes: 5_000,
      estimatedSeconds: 600,
    })
  })
})
