import type { ReactElement, ReactNode } from 'react'
import { CircleAlert, CloudUpload, Info } from 'lucide-react'
import { usePublishPlan } from '../../api/publish-queries'
import type { PublishPlan } from '../../api/types'
import { DialogBody, DialogFooter } from '../../app/Dialog'
import { Button, ButtonVariant } from '../../ui/Button'
import { StatTile } from '../../ui/StatTile'
import { formatBytes } from '../compression/format'

const NOTE = 'flex gap-2.5 rounded-xl bg-page px-4 py-3 text-[13px]'
const NOTE_ICON = 'mt-0.5 size-4 shrink-0'

interface PublishPlanSummaryProps {
  readonly courseId: string
  readonly isStarting: boolean
  readonly onStart: () => void
  readonly onDismiss: () => void
  readonly children: ReactNode
}

export function PublishPlanSummary({
  courseId,
  isStarting,
  onStart,
  onDismiss,
  children,
}: PublishPlanSummaryProps): ReactElement {
  const plan = usePublishPlan(courseId)
  const canStart = plan.data?.fits === true && !isStarting

  return (
    <>
      <DialogBody>
        {children}
        {plan.isError && (
          <p className="text-danger" role="alert">
            Không đọc được bucket ({plan.error.message}).
          </p>
        )}
        {plan.isPending && <p className="text-muted">Đang xem khoá học và bucket…</p>}
        {plan.data !== undefined && <PlanDetails plan={plan.data} courseId={courseId} />}
      </DialogBody>
      <DialogFooter>
        <Button onClick={onDismiss}>Thôi</Button>
        <Button
          variant={ButtonVariant.Primary}
          icon={CloudUpload}
          disabled={!canStart}
          onClick={onStart}
        >
          Đưa lên R2
        </Button>
      </DialogFooter>
    </>
  )
}

interface PlanDetailsProps {
  readonly plan: PublishPlan
  readonly courseId: string
}

function PlanDetails({ plan, courseId }: PlanDetailsProps): ReactElement {
  const isAlreadyThere = plan.published?.courseId === courseId
  return (
    <>
      <div className="flex flex-wrap gap-3">
        <StatTile label="Số file" value={String(plan.files)} />
        <StatTile
          label={`Dung lượng (hạn mức ${formatBytes(plan.limitBytes)})`}
          value={formatBytes(plan.totalBytes)}
        />
      </div>
      {!plan.fits && (
        <p className={`${NOTE} text-danger`} role="alert">
          <CircleAlert className={NOTE_ICON} aria-hidden="true" />
          Khoá nặng hơn hạn mức của gói miễn phí. Nén video trước rồi thử lại.
        </p>
      )}
      {plan.replaces && plan.published !== undefined && (
        <p className={`${NOTE} text-muted`}>
          <CircleAlert className={`${NOTE_ICON} text-danger`} aria-hidden="true" />
          Khoá «{plan.published.title}» đang trên R2 sẽ bị xoá trước. Bản trên VPS sẽ trống cho đến
          khi đưa lên xong.
        </p>
      )}
      {isAlreadyThere && (
        <p className={`${NOTE} text-muted`}>
          <Info className={`${NOTE_ICON} text-accent-soft`} aria-hidden="true" />
          Khoá này đã ở trên R2; chỉ gửi lại các file mới hoặc đã đổi.
        </p>
      )}
    </>
  )
}
