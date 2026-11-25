import { useRef, useState, type ReactElement } from 'react'
import { Link, useSearchParams } from 'react-router'
import { Search, SearchX, X } from 'lucide-react'
import { useSubtitleSearch } from '../../api/queries'
import type { SearchHit } from '../../api/types'
import { SEARCH_QUERY_PARAM, lectureAtPath } from '../../app/paths'
import { MS_PER_SECOND, formatTimestamp } from '../../app/timestamp'
import { EmptyState } from '../../ui/EmptyState'
import { resultCountText } from './result-count'
import { splitSnippet } from './snippet'
import { useDebouncedValue } from './use-debounced-value'

const TYPING_PAUSE_MS = 250
const PILL = 'shrink-0 rounded-full px-2 py-[3px] text-xs'

export function SearchPage(): ReactElement {
  const [searchParams, setSearchParams] = useSearchParams()
  const [typed, setTyped] = useState(searchParams.get(SEARCH_QUERY_PARAM) ?? '')
  const inputRef = useRef<HTMLInputElement | null>(null)
  const query = useDebouncedValue(typed.trim(), TYPING_PAUSE_MS)
  const hits = useSubtitleSearch(query)
  const changeTyped = (value: string): void => {
    setTyped(value)
    setSearchParams({ [SEARCH_QUERY_PARAM]: value }, { replace: true })
  }

  return (
    <main className="grid max-w-[1096px] content-start gap-4 px-5 py-8 md:px-10 md:py-10">
      <h1 className="text-[30px] font-semibold">Tìm trong phụ đề</h1>
      <label className="flex h-[52px] items-center gap-2.5 rounded-xl border border-line bg-input pr-3.5 pl-4 focus-within:border-accent">
        <Search className="size-[18px] shrink-0 text-muted" aria-hidden="true" />
        <input
          ref={inputRef}
          className="h-full min-w-0 flex-1 rounded-none border-0 bg-transparent text-[15px] outline-none placeholder:text-muted [&::-webkit-search-cancel-button]:appearance-none"
          type="search"
          autoFocus
          aria-label="Tìm trong phụ đề"
          placeholder="Gõ một cụm từ, có dấu hay không dấu đều được"
          value={typed}
          onChange={(event) => {
            changeTyped(event.target.value)
          }}
        />
        {typed !== '' && (
          <button
            type="button"
            className="grid size-7 shrink-0 place-items-center rounded-md bg-transparent p-0 text-muted hover:text-ink"
            aria-label="Xoá ô tìm"
            onClick={() => {
              changeTyped('')
              inputRef.current?.focus()
            }}
          >
            <X className="size-[18px]" aria-hidden="true" />
          </button>
        )}
      </label>
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
    return (
      <p className="text-sm text-danger" role="alert">
        Không tìm được: {error.message}
      </p>
    )
  if (hits === undefined)
    return (
      <p className="text-[13px] text-muted" role="status">
        Đang tìm…
      </p>
    )
  if (hits.length === 0)
    return <EmptyState icon={SearchX} title={`Không có câu nào chứa “${query}”.`} />
  return (
    <>
      <p className="text-[13px] text-muted" role="status">
        {resultCountText(hits.length, query)}
      </p>
      <ol className="stagger-children grid gap-2.5">
        {hits.map((hit) => (
          <li key={`${hit.lectureId}-${hit.language}-${String(hit.startMs)}-${hit.snippet}`}>
            <SearchHitCard hit={hit} />
          </li>
        ))}
      </ol>
    </>
  )
}

interface SearchHitCardProps {
  readonly hit: SearchHit
}

function SearchHitCard({ hit }: SearchHitCardProps): ReactElement {
  return (
    <Link
      to={lectureAtPath(hit.courseId, hit.lectureId, Math.floor(hit.startMs / MS_PER_SECOND))}
      className="grid gap-2 rounded-xl bg-card px-4 py-3.5 hover:bg-ink/5"
    >
      <span className="flex items-center gap-3">
        <span className="min-w-0 flex-1 truncate text-xs text-muted">
          {hit.courseTitle} › {hit.lectureTitle}
        </span>
        <span className={`${PILL} bg-accent/20 text-accent-soft tabular-nums`}>
          {formatTimestamp(hit.startMs)}
        </span>
        <span className={`${PILL} bg-surface-raised text-muted`}>{hit.language}</span>
      </span>
      <span className="text-sm">
        {splitSnippet(hit.snippet).map((part, position) =>
          part.isMatch ? (
            <mark key={position} className="bg-transparent font-semibold text-accent-soft">
              {part.text}
            </mark>
          ) : (
            <span key={position}>{part.text}</span>
          ),
        )}
      </span>
    </Link>
  )
}
