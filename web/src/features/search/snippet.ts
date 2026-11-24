// The server marks each match in a snippet with these private-use characters.
const HIGHLIGHT_START = '\ue000'
const HIGHLIGHT_END = '\ue001'

export interface SnippetPart {
  readonly text: string
  readonly isMatch: boolean
}

export function splitSnippet(snippet: string): readonly SnippetPart[] {
  const parts: SnippetPart[] = []
  for (const [position, piece] of snippet.split(HIGHLIGHT_START).entries()) {
    const [match, rest] = position === 0 ? ['', piece] : splitOnce(piece, HIGHLIGHT_END)
    if (match !== '') parts.push({ text: match, isMatch: true })
    if (rest !== '') parts.push({ text: rest, isMatch: false })
  }
  return parts
}

function splitOnce(text: string, separator: string): readonly [string, string] {
  const at = text.indexOf(separator)
  return at < 0 ? [text, ''] : [text.slice(0, at), text.slice(at + separator.length)]
}
