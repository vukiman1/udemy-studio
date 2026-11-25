import type { ReactElement } from 'react'
import { CircleCheck, CircleStop } from 'lucide-react'
import type { CompressionRun } from '../../api/types'
import { formatBytes } from './format'

interface CompressionResultProps {
  readonly run: CompressionRun
}

export function CompressionResult({ run }: CompressionResultProps): ReactElement {
  const compressed = run.finished - run.skipped.length - run.failed.length
  const isCancelled = run.state === 'cancelled'
  const Icon = isCancelled ? CircleStop : CircleCheck
  return (
    <div className="grid gap-2 text-[13px]">
      <div
        className={`flex gap-3 rounded-xl px-4 py-3.5 ${isCancelled ? 'bg-page' : 'bg-success-soft'}`}
      >
        <Icon
          className={`mt-0.5 size-[18px] shrink-0 ${isCancelled ? 'text-muted' : 'text-success'}`}
          aria-hidden="true"
        />
        <div className="grid gap-0.5">
          <p className={`text-[15px] font-semibold ${isCancelled ? '' : 'text-success'}`}>
            {isCancelled ? 'Đã dừng' : 'Nén xong'} {compressed}/{run.total} video
          </p>
          <p>
            {compressed > 0
              ? `${formatBytes(run.bytesBefore)} → ${formatBytes(run.bytesAfter)} · tiết kiệm ${formatBytes(run.bytesBefore - run.bytesAfter)}`
              : 'Chưa có video nào được thay.'}
          </p>
        </div>
      </div>
      {run.skipped.map((outcome) => (
        <p key={outcome.path} className="text-muted">
          Giữ bản gốc: {outcome.lectureTitle} ({outcome.reason})
        </p>
      ))}
      {run.failed.map((outcome) => (
        <p key={outcome.path} className="text-danger">
          Lỗi: {outcome.lectureTitle} ({outcome.reason})
        </p>
      ))}
    </div>
  )
}
