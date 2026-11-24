import type { ReactElement } from 'react'
import { useTranslationPlan } from '../../api/translation-queries'
import { formatUsd } from './format-usd'

interface TranslationPlanSummaryProps {
  readonly courseId: string
  readonly isStarting: boolean
  readonly onStart: () => void
  readonly onDismiss: () => void
}

export function TranslationPlanSummary({
  courseId,
  isStarting,
  onStart,
  onDismiss,
}: TranslationPlanSummaryProps): ReactElement {
  const plan = useTranslationPlan(courseId)
  if (plan.isError)
    return (
      <p className="m-0 text-danger" role="alert">
        Không lập được kế hoạch dịch ({plan.error.message}).
      </p>
    )
  if (plan.data === undefined) return <p className="m-0 text-muted">Đang đọc phụ đề…</p>

  const { lectures, alreadyTranslated, estimatedUsd } = plan.data
  return (
    <div className="grid gap-2">
      <p className="m-0">
        {lectures === 0
          ? 'Không còn bài nào cần dịch.'
          : `Sẽ dịch ${String(lectures)} bài, ước tính khoảng ${formatUsd(estimatedUsd)}.`}
      </p>
      {alreadyTranslated > 0 && (
        <p className="m-0 text-[0.9rem] text-muted">
          Bỏ qua {alreadyTranslated} bài đã có phụ đề tiếng Việt.
        </p>
      )}
      <div className="flex gap-3">
        {lectures > 0 && (
          <button type="button" disabled={isStarting} onClick={onStart}>
            Bắt đầu dịch
          </button>
        )}
        <button type="button" onClick={onDismiss}>
          Thôi
        </button>
      </div>
    </div>
  )
}
