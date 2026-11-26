import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { ArrowRight, History } from 'lucide-react'
import { useHistory } from '../../api/queries'
import type { HistoryEntry } from '../../api/types'
import { lectureAtPath, lecturePath } from '../../app/paths'
import { MS_PER_SECOND, formatTimestamp } from '../../app/timestamp'
import { ButtonSize, ButtonVariant, buttonClass } from '../../ui/Button'
import { EmptyState } from '../../ui/EmptyState'

export function HistoryPage(): ReactElement {
  const history = useHistory()

  return (
    <main className="grid max-w-[1136px] content-start gap-7 px-5 py-8 md:px-10 md:py-10">
      <h1 className="text-[30px] font-semibold">History</h1>
      {history.isError && (
        <p className="text-sm text-danger" role="alert">
          Could not load history ({history.error.message}).
        </p>
      )}
      {history.isPending && <p className="text-sm text-muted">Loading history…</p>}
      {history.data?.length === 0 && (
        <EmptyState
          icon={History}
          title="Nothing watched yet"
          body="Open a course from the library to start."
        />
      )}
      {history.data !== undefined && history.data.length > 0 && (
        <ol className="stagger-children grid gap-5">
          {history.data.map((entry) => (
            <li key={`${entry.courseId}/${entry.lectureId}`}>
              <HistoryCard entry={entry} />
            </li>
          ))}
        </ol>
      )}
    </main>
  )
}

interface HistoryCardProps {
  readonly entry: HistoryEntry
}

function HistoryCard({ entry }: HistoryCardProps): ReactElement {
  const resumeTo = entry.completed
    ? lecturePath(entry.courseId, entry.lectureId)
    : lectureAtPath(entry.courseId, entry.lectureId, Math.floor(entry.positionSeconds))
  return (
    <article className="flex flex-wrap items-center gap-x-6 gap-y-3 rounded-[22px] bg-card px-5 py-6 md:px-[30px] md:py-9">
      <div className="grid min-w-0 flex-1 basis-64 gap-1.5">
        <h2 className="text-xl font-semibold wrap-anywhere">{entry.lectureTitle}</h2>
        <p className="text-sm wrap-anywhere text-muted">{entry.courseTitle}</p>
      </div>
      <div className="flex items-center justify-between gap-6 max-md:w-full">
        <p
          className={`text-[13px] font-semibold md:text-[15px] ${entry.completed ? 'text-success' : 'text-muted'}`}
        >
          {entry.completed
            ? 'Completed'
            : `In progress · stopped at ${formatTimestamp(entry.positionSeconds * MS_PER_SECOND)}`}
        </p>
        <Link
          to={resumeTo}
          className={`${buttonClass(ButtonVariant.Ghost, ButtonSize.Large)} max-md:-mr-5`}
        >
          <ArrowRight aria-hidden="true" />
          Continue
        </Link>
      </div>
    </article>
  )
}
