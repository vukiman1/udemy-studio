import type { ReactElement } from 'react'
import type { CompressionRun } from '../../api/types'
import { formatBytes } from './format'

interface CompressionResultProps {
  readonly run: CompressionRun
}

export function CompressionResult({ run }: CompressionResultProps): ReactElement {
  const compressed = run.finished - run.skipped.length - run.failed.length
  return (
    <div className="grid gap-1 text-[0.9rem]">
      <p className="m-0">
        {run.state === 'cancelled' ? 'Đã dừng. ' : 'Nén xong. '}
        {compressed > 0
          ? `${String(compressed)} video: ${formatBytes(run.bytesBefore)} → ${formatBytes(run.bytesAfter)}, tiết kiệm ${formatBytes(run.bytesBefore - run.bytesAfter)}.`
          : 'Chưa có video nào được thay.'}
      </p>
      {run.skipped.map((outcome) => (
        <p key={outcome.path} className="m-0 text-muted">
          Giữ bản gốc: {outcome.lectureTitle} ({outcome.reason})
        </p>
      ))}
      {run.failed.map((outcome) => (
        <p key={outcome.path} className="m-0 text-danger">
          Lỗi: {outcome.lectureTitle} ({outcome.reason})
        </p>
      ))}
    </div>
  )
}
