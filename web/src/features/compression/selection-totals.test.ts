import { describe, expect, it } from 'vitest'
import type { CompressionChapter } from '../../api/types'
import { estimatedSizeAfter, selectionTotals } from './selection-totals'

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

describe('estimatedSizeAfter', () => {
  it('ranges from a fifth to a half of the current size before anything was learned', () => {
    expect(estimatedSizeAfter(1_000)).toEqual({ lowBytes: 200, highBytes: 500 })
  })

  it('adds the likely size once earlier encodes taught it a share', () => {
    expect(estimatedSizeAfter(1_000, 0.35)).toEqual({
      lowBytes: 200,
      highBytes: 500,
      likelyBytes: 350,
    })
  })

  it('widens the range to hold a learned share outside it', () => {
    expect(estimatedSizeAfter(1_000, 0.8)).toEqual({
      lowBytes: 200,
      highBytes: 800,
      likelyBytes: 800,
    })
  })
})
