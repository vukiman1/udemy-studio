import type { ReactElement, ReactNode } from 'react'
import { Archive, Info } from 'lucide-react'
import { useCompressionPlan } from '../../api/compression-queries'
import { DialogBody, DialogFooter } from '../../app/Dialog'
import { Button, ButtonVariant } from '../../ui/Button'
import { StatTile } from '../../ui/StatTile'
import { formatBytes, formatRoughDuration } from './format'

interface CompressionPlanSummaryProps {
  readonly courseId: string
  readonly isStarting: boolean
  readonly onStart: () => void
  readonly onDismiss: () => void
  readonly children: ReactNode
}

export function CompressionPlanSummary({
  courseId,
  isStarting,
  onStart,
  onDismiss,
  children,
}: CompressionPlanSummaryProps): ReactElement {
  const plan = useCompressionPlan(courseId)
  const files = plan.data?.files ?? 0

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
            {files === 0 ? (
              <p>Nothing left to compress.</p>
            ) : (
              <>
                <div className="flex flex-wrap gap-3">
                  <StatTile label="Videos" value={String(files)} />
                  <StatTile label="Current size" value={formatBytes(plan.data.totalBytes)} />
                  <StatTile
                    label="Estimated time"
                    value={`~ ${formatRoughDuration(plan.data.estimatedSeconds)}`}
                  />
                </div>
                <p className="flex gap-2.5 rounded-xl bg-page px-4 py-3 text-[13px] text-muted">
                  <Info className="mt-0.5 size-4 shrink-0 text-accent-soft" aria-hidden="true" />
                  Playback keeps working; reload if the lecture being replaced stutters.
                </p>
              </>
            )}
            <SkippedVideos {...plan.data} />
          </>
        )}
      </DialogBody>
      <DialogFooter>
        {files > 0 ? (
          <>
            <Button onClick={onDismiss}>Cancel</Button>
            <Button
              variant={ButtonVariant.Primary}
              icon={Archive}
              disabled={isStarting}
              onClick={onStart}
            >
              Compress
            </Button>
          </>
        ) : (
          <Button variant={ButtonVariant.Primary} onClick={onDismiss}>
            Close
          </Button>
        )}
      </DialogFooter>
    </>
  )
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
