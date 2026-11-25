const MAX_SEARCH_HITS = 50

export function resultCountText(count: number, query: string): string {
  const shown =
    count >= MAX_SEARCH_HITS ? `${String(count)} kết quả đầu tiên` : `${String(count)} kết quả`
  return `${shown} cho “${query}”`
}
