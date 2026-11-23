import type { Cue, CueLine } from '../../api/types'

const LONGEST_CUE_MS = 60_000

export function activeLinesAt(
  cues: readonly Cue[],
  timeMs: number,
  languageOrder: readonly string[],
): CueLine[] {
  const active: Cue[] = []
  for (let index = firstStartingAfter(cues, timeMs) - 1; index >= 0; index--) {
    const cue = cues[index]
    if (cue === undefined || cue.startMs < timeMs - LONGEST_CUE_MS) break
    if (timeMs < cue.endMs) active.push(cue)
  }
  const rank = (line: CueLine): number => {
    const position = languageOrder.indexOf(line.language)
    return position < 0 ? languageOrder.length : position
  }
  return active
    .reverse()
    .flatMap((cue) => cue.lines)
    .sort((left, right) => rank(left) - rank(right))
}

function firstStartingAfter(cues: readonly Cue[], timeMs: number): number {
  let low = 0
  let high = cues.length
  while (low < high) {
    const middle = (low + high) >> 1
    const cue = cues[middle]
    if (cue !== undefined && cue.startMs <= timeMs) low = middle + 1
    else high = middle
  }
  return low
}
