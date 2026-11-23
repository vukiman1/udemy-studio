import type { ReactElement } from 'react'

interface StaleDataNoticeProps {
  readonly error: Error
}

export function StaleDataNotice({ error }: StaleDataNoticeProps): ReactElement {
  return (
    <p className="m-0 px-4 py-1.5 text-[0.85rem] text-danger" role="status">
      Không làm mới được dữ liệu ({error.message}), đang hiển thị bản đã tải.
    </p>
  )
}
