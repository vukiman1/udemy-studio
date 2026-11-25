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
  // What the dialog shows above the plan, such as the last run's result.
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
            Không kiểm tra được video ({plan.error.message}).
          </p>
        )}
        {plan.isPending && <p className="text-muted">Đang kiểm tra các video…</p>}
        {plan.data !== undefined && (
          <>
            {files === 0 ? (
              <p>Không còn video nào cần nén.</p>
            ) : (
              <>
                <div className="flex flex-wrap gap-3">
                  <StatTile label="Video sẽ nén" value={String(files)} />
                  <StatTile label="Dung lượng hiện tại" value={formatBytes(plan.data.totalBytes)} />
                  <StatTile
                    label="Thời gian ước tính"
                    value={`~ ${formatRoughDuration(plan.data.estimatedSeconds)}`}
                  />
                </div>
                <p className="flex gap-2.5 rounded-xl bg-page px-4 py-3 text-[13px] text-muted">
                  <Info className="mt-0.5 size-4 shrink-0 text-accent-soft" aria-hidden="true" />
                  Trong lúc nén vẫn học được. Bài đang được nén có thể giật một lần lúc file được
                  thay, tải lại trang là hết.
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
            <Button onClick={onDismiss}>Thôi</Button>
            <Button
              variant={ButtonVariant.Primary}
              icon={Archive}
              disabled={isStarting}
              onClick={onStart}
            >
              Bắt đầu nén
            </Button>
          </>
        ) : (
          <Button variant={ButtonVariant.Primary} onClick={onDismiss}>
            Đóng
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
    alreadyHevc > 0 ? `${String(alreadyHevc)} video đã là H.265` : '',
    otherCodec > 0 ? `${String(otherCodec)} video mã hoá khác` : '',
    notWorthIt > 0 ? `${String(notWorthIt)} video lần trước nén không lợi hơn` : '',
    unreadable > 0 ? `${String(unreadable)} video không đọc được (có thể tải chưa xong)` : '',
  ].filter((part) => part !== '')
  if (leftAlone.length === 0) return null
  return <p className="text-xs text-muted">Bỏ qua: {leftAlone.join(', ')}.</p>
}
