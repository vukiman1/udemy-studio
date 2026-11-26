import { useState, type ReactElement, type ReactNode } from 'react'
import { CircleAlert, Cloud, CloudUpload } from 'lucide-react'
import { usePublishPlan } from '../../api/publish-queries'
import type { PublishPlan } from '../../api/types'
import { DialogBody, DialogFooter } from '../../app/Dialog'
import { Button, ButtonVariant } from '../../ui/Button'
import { ChapterPicker } from '../../ui/ChapterPicker'
import { StatTile } from '../../ui/StatTile'
import { formatBytes } from '../compression/format'
import { hasChanges, initialSelection, selectedBytes } from './chapter-selection'

const NOTE = 'flex gap-2.5 rounded-xl bg-page px-4 py-3 text-[13px]'
const NOTE_ICON = 'mt-0.5 size-4 shrink-0'

interface PublishPlanSummaryProps {
  readonly courseId: string
  readonly isStarting: boolean
  readonly onStart: (chapters: readonly string[]) => void
  readonly onDismiss: () => void
  readonly removeButton: ReactNode
  readonly children: ReactNode
}

export function PublishPlanSummary({
  courseId,
  isStarting,
  onStart,
  onDismiss,
  removeButton,
  children,
}: PublishPlanSummaryProps): ReactElement {
  const plan = usePublishPlan(courseId)
  if (plan.data !== undefined)
    return (
      <PublishPlanForm
        key={plan.data.published?.publishedAt ?? ''}
        plan={plan.data}
        courseId={courseId}
        isStarting={isStarting}
        onStart={onStart}
        onDismiss={onDismiss}
        removeButton={removeButton}
      >
        {children}
      </PublishPlanForm>
    )

  return (
    <>
      <DialogBody>
        {children}
        {plan.isError ? (
          <p className="text-danger" role="alert">
            {plan.error.message}
          </p>
        ) : (
          <p className="text-muted">Loading…</p>
        )}
      </DialogBody>
      <DialogFooter>
        <Button onClick={onDismiss}>Cancel</Button>
      </DialogFooter>
    </>
  )
}

interface PublishPlanFormProps extends PublishPlanSummaryProps {
  readonly plan: PublishPlan
}

function PublishPlanForm({
  plan,
  courseId,
  isStarting,
  onStart,
  onDismiss,
  removeButton,
  children,
}: PublishPlanFormProps): ReactElement {
  const [selected, setSelected] = useState(() => initialSelection(plan, courseId))
  const totalBytes = selectedBytes(plan, selected)
  const fits = totalBytes <= plan.limitBytes
  const isOnR2 = plan.published?.courseId === courseId

  return (
    <>
      <DialogBody>
        {children}
        <div className="flex flex-wrap gap-3">
          <StatTile
            label="Chapters"
            value={`${String(selected.size)}/${String(plan.chapters.length)}`}
          />
          <StatTile
            label={`Size (≤ ${formatBytes(plan.limitBytes)})`}
            value={formatBytes(totalBytes)}
          />
        </div>
        <ChapterPicker
          chapters={plan.chapters.map((chapter) => ({
            dir: chapter.dir,
            detail: (
              <>
                {chapter.published && (
                  <Cloud className="size-3.5 text-accent-soft" aria-label="On R2" />
                )}
                {formatBytes(chapter.sizeBytes)}
              </>
            ),
          }))}
          selected={selected}
          onChange={setSelected}
        />
        {!fits && (
          <p className={`${NOTE} text-danger`} role="alert">
            <CircleAlert className={NOTE_ICON} aria-hidden="true" />
            Over the limit
          </p>
        )}
        {plan.replaces && plan.published !== undefined && (
          <p className={`${NOTE} text-muted`}>
            <CircleAlert className={`${NOTE_ICON} text-danger`} aria-hidden="true" />
            Replaces «{plan.published.title}» on R2
          </p>
        )}
      </DialogBody>
      <DialogFooter>
        <Button onClick={onDismiss}>Cancel</Button>
        {isOnR2 && selected.size === 0 ? (
          removeButton
        ) : (
          <Button
            variant={ButtonVariant.Primary}
            icon={CloudUpload}
            disabled={
              !fits || selected.size === 0 || isStarting || !hasChanges(plan, courseId, selected)
            }
            onClick={() => {
              onStart(
                plan.chapters.map((chapter) => chapter.dir).filter((dir) => selected.has(dir)),
              )
            }}
          >
            {isOnR2 ? 'Update' : 'Publish'}
          </Button>
        )}
      </DialogFooter>
    </>
  )
}
