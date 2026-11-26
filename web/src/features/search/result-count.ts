const MAX_SEARCH_HITS = 50

export function resultCountText(count: number, query: string): string {
  const shown =
    count >= MAX_SEARCH_HITS ? `First ${String(count)} results` : `${String(count)} results`
  return `${shown} for “${query}”`
}
