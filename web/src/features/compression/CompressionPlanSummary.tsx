import type { ReactElement } from 'react'
import { useCompressionPlan } from '../../api/compression-queries'
import { formatBytes, formatRoughDuration } from './format'

interface CompressionPlanSummaryProps {
  readonly courseId: string
  readonly isStarting: boolean
  readonly onStart: () => void
  readonly onDismiss: () => void
}

export function CompressionPlanSummary({
  courseId,
  isStarting,
  onStart,
  onDismiss,
}: CompressionPlanSummaryProps): ReactElement {
  const plan = useCompressionPlan(courseId)
  if (plan.isError)
    return (
      <p className="m-0 text-danger" role="alert">
        Không kiểm tra được video ({plan.error.message}).
      </p>
    )
  if (plan.data === undefined) return <p className="m-0 text-muted">Đang kiểm tra các video…</p>

  const { files, totalBytes, estimatedSeconds, alreadyHevc, otherCodec, notWorthIt, unreadable } =
    plan.data
  const leftAlone = [
    alreadyHevc > 0 ? `${String(alreadyHevc)} video đã là H.265` : '',
    otherCodec > 0 ? `${String(otherCodec)} video mã hoá khác` : '',
    notWorthIt > 0 ? `${String(notWorthIt)} video lần trước nén không lợi hơn` : '',
    unreadable > 0 ? `${String(unreadable)} video không đọc được (có thể tải chưa xong)` : '',
  ].filter((part) => part !== '')

  return (
    <div className="grid gap-2">
      {files === 0 ? (
        <p className="m-0">Không còn video nào cần nén.</p>
      ) : (
        <p className="m-0">
          Sẽ nén {files} video H.264 ({formatBytes(totalBytes)}), mất khoảng{' '}
          {formatRoughDuration(estimatedSeconds)}. Trong lúc nén vẫn học được; bài đang được nén có
          thể giật một lần lúc file được thay, tải lại trang là hết.
        </p>
      )}
      {leftAlone.length > 0 && (
        <p className="m-0 text-[0.9rem] text-muted">Bỏ qua: {leftAlone.join(', ')}.</p>
      )}
      <div className="flex gap-3">
        {files > 0 && (
          <button type="button" disabled={isStarting} onClick={onStart}>
            Bắt đầu nén
          </button>
        )}
        <button type="button" onClick={onDismiss}>
          Thôi
        </button>
      </div>
    </div>
  )
}
