import { useState, type ReactElement } from 'react'
import { Link, useSearchParams } from 'react-router'
import { useSubtitleSearch } from '../../api/queries'
import type { SearchHit } from '../../api/types'
import { PageStatus } from '../../app/PageStatus'
import { SEARCH_QUERY_PARAM, lectureAtPath } from '../../app/paths'
import { splitSnippet } from './snippet'
import { MS_PER_SECOND, formatTimestamp } from './timestamp'
import { useDebouncedValue } from './use-debounced-value'

const TYPING_PAUSE_MS = 250

export function SearchPage(): ReactElement {
  const [searchParams, setSearchParams] = useSearchParams()
  const [typed, setTyped] = useState(searchParams.get(SEARCH_QUERY_PARAM) ?? '')
  const query = useDebouncedValue(typed.trim(), TYPING_PAUSE_MS)
  const hits = useSubtitleSearch(query)

  return (
    <main className="mx-auto grid max-w-[900px] gap-4 px-4 py-6">
      <header className="flex flex-wrap items-center gap-3">
        <Link to="/" className="text-accent underline">
          ← Thư viện
        </Link>
        <h1 className="text-2xl font-bold">Tìm trong phụ đề</h1>
      </header>
      <input
        className="rounded-md border border-line bg-page px-3 py-2"
        type="search"
        autoFocus
        placeholder="Gõ một cụm từ, có dấu hay không dấu đều được"
        value={typed}
        onChange={(event) => {
          setTyped(event.target.value)
          setSearchParams({ [SEARCH_QUERY_PARAM]: event.target.value }, { replace: true })
        }}
      />
      {query !== '' && <SearchResults query={query} hits={hits.data} error={hits.error} />}
    </main>
  )
}

interface SearchResultsProps {
  readonly query: string
  readonly hits: readonly SearchHit[] | undefined
  readonly error: Error | null
}

function SearchResults({ query, hits, error }: SearchResultsProps): ReactElement {
  if (error !== null)
    return <PageStatus tone="error" message={`Không tìm được: ${error.message}`} />
  if (hits === undefined) return <PageStatus tone="info" message="Đang tìm…" />
  if (hits.length === 0)
    return <PageStatus tone="info" message={`Không có câu nào chứa "${query}".`} />
  return (
    <ol className="grid gap-2">
      {hits.map((hit) => (
        <li key={`${hit.lectureId}-${hit.language}-${String(hit.startMs)}-${hit.snippet}`}>
          <Link
            to={lectureAtPath(hit.courseId, hit.lectureId, Math.floor(hit.startMs / MS_PER_SECOND))}
            className="grid gap-1 rounded-lg border border-line bg-surface p-3 hover:border-accent"
          >
            <span className="text-[0.85rem] text-muted">
              {hit.courseTitle} › {hit.lectureTitle} · {formatTimestamp(hit.startMs)} ·{' '}
              {hit.language}
            </span>
            <span>
              {splitSnippet(hit.snippet).map((part, position) =>
                part.isMatch ? (
                  <mark key={position} className="rounded bg-accent/30 text-inherit">
                    {part.text}
                  </mark>
                ) : (
                  <span key={position}>{part.text}</span>
                ),
              )}
            </span>
          </Link>
        </li>
      ))}
    </ol>
  )
}
