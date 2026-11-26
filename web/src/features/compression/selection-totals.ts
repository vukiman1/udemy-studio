import type { CompressionChapter } from '../../api/types'

const SMALLEST_SHARE_AFTER = 0.2
const LARGEST_SHARE_AFTER = 0.5

export interface SelectionTotals {
  readonly videos: number
  readonly totalBytes: number
  readonly estimatedSeconds: number
}

export function selectionTotals(
  chapters: readonly CompressionChapter[],
  selected: ReadonlySet<string>,
): SelectionTotals {
  return chapters
    .filter((chapter) => selected.has(chapter.dir))
    .reduce<SelectionTotals>(
      (totals, chapter) => ({
        videos: totals.videos + chapter.videos,
        totalBytes: totals.totalBytes + chapter.totalBytes,
        estimatedSeconds: totals.estimatedSeconds + chapter.estimatedSeconds,
      }),
      { videos: 0, totalBytes: 0, estimatedSeconds: 0 },
    )
}

export interface SizeEstimate {
  readonly lowBytes: number
  readonly highBytes: number
  readonly likelyBytes?: number
}

export function estimatedSizeAfter(totalBytes: number, learnedShare?: number): SizeEstimate {
  const range = {
    lowBytes: totalBytes * Math.min(SMALLEST_SHARE_AFTER, learnedShare ?? SMALLEST_SHARE_AFTER),
    highBytes: totalBytes * Math.max(LARGEST_SHARE_AFTER, learnedShare ?? LARGEST_SHARE_AFTER),
  }
  return learnedShare === undefined ? range : { ...range, likelyBytes: totalBytes * learnedShare }
}
