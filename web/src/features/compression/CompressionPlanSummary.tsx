import { useState, type ReactElement, type ReactNode } from 'react'
import { Archive } from 'lucide-react'
import { useCompressionPlan } from '../../api/compression-queries'
import type { CompressionPlan } from '../../api/types'
import { DialogBody, DialogFooter } from '../../app/Dialog'
import { Button, ButtonVariant } from '../../ui/Button'
import { ChapterPicker } from '../../ui/ChapterPicker'
import { StatTile } from '../../ui/StatTile'
import { formatBytes, formatRoughDuration } from './format'
import { selectionTotals } from './selection-totals'

const NOTHING_SELECTED = '—'

interface CompressionPlanSummaryProps {
  readonly courseId: string
  readonly isStarting: boolean
  readonly onStart: (chapters: readonly string[]) => void
  readonly onDismiss: () => void
  readonly children: ReactNode
}

export function CompressionPlanSummary({
  courseId,
  children,
  ...actions
}: CompressionPlanSummaryProps): ReactElement {
  const plan = useCompressionPlan(courseId)
  if (plan.data !== undefined && plan.data.chapters.length > 0)
    return (
      <CompressionPlanForm
        key={plan.data.chapters.map((chapter) => chapter.dir).join('|')}
        plan={plan.data}
        {...actions}
      >
        {children}
      </CompressionPlanForm>
    )

  return (
    <>
      <DialogBody>
        {children}
        {plan.isError && (
          <p className="text-danger" role="alert">
            Could not check the videos ({plan.error.message}).
          </p>
        )}
        {plan.isPending && <p className="text-muted">Checking videos…</p>}
        {plan.data !== undefined && (
          <>
            <p>Nothing left to compress.</p>
            <SkippedVideos {...plan.data} />
          </>
        )}
      </DialogBody>
      <DialogFooter>
        <Button variant={ButtonVariant.Primary} onClick={actions.onDismiss}>
          Close
        </Button>
      </DialogFooter>
    </>
  )
}

interface CompressionPlanFormProps extends Omit<CompressionPlanSummaryProps, 'courseId'> {
  readonly plan: CompressionPlan
}

function CompressionPlanForm({
  plan,
  isStarting,
  onStart,
  onDismiss,
  children,
}: CompressionPlanFormProps): ReactElement {
  const [selected, setSelected] = useState<ReadonlySet<string>>(
    () => new Set(plan.chapters.map((chapter) => chapter.dir)),
  )
  const totals = selectionTotals(plan.chapters, selected)

  return (
    <>
      <DialogBody>
        {children}
        <div className="flex flex-wrap gap-3">
          <StatTile label="Videos" value={String(totals.videos)} />
          <StatTile label="Current size" value={formatBytes(totals.totalBytes)} />
          <StatTile
            label="Estimated time"
            value={
              totals.videos > 0
                ? `~ ${formatRoughDuration(totals.estimatedSeconds)}`
                : NOTHING_SELECTED
            }
          />
        </div>
        <ChapterPicker
          chapters={plan.chapters.map((chapter) => ({
            dir: chapter.dir,
            detail: `${videoCount(chapter.videos)} · ${formatBytes(chapter.totalBytes)}`,
          }))}
          selected={selected}
          onChange={setSelected}
        />
        <SkippedVideos {...plan} />
      </DialogBody>
      <DialogFooter>
        <Button onClick={onDismiss}>Cancel</Button>
        <Button
          variant={ButtonVariant.Primary}
          icon={Archive}
          disabled={isStarting || selected.size === 0}
          onClick={() => {
            onStart([...selected])
          }}
        >
          Compress
        </Button>
      </DialogFooter>
    </>
  )
}

function videoCount(videos: number): string {
  return videos === 1 ? '1 video' : `${String(videos)} videos`
}

interface SkippedVideosProps {
  readonly alreadyHevc: number
  readonly otherCodec: number
  readonly notWorthIt: number
  readonly unreadable: number
}

function SkippedVideos({
  alreadyHevc,
  otherCodec,
  notWorthIt,
  unreadable,
}: SkippedVideosProps): ReactElement | null {
  const leftAlone = [
    alreadyHevc > 0 ? `${String(alreadyHevc)} already H.265` : '',
    otherCodec > 0 ? `${String(otherCodec)} other codec` : '',
    notWorthIt > 0 ? `${String(notWorthIt)} not smaller last time` : '',
    unreadable > 0 ? `${String(unreadable)} unreadable (maybe still downloading)` : '',
  ].filter((part) => part !== '')
  if (leftAlone.length === 0) return null
  return <p className="text-xs text-muted">Skipped: {leftAlone.join(', ')}.</p>
}
