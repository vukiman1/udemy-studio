import type { CompressionChapter } from '../../api/types'

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
