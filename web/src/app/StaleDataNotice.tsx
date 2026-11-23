import type { ReactElement } from 'react'
import styles from './StaleDataNotice.module.css'

interface StaleDataNoticeProps {
  readonly error: Error
}

export function StaleDataNotice({ error }: StaleDataNoticeProps): ReactElement {
  return (
    <p className={styles.notice} role="status">
      Không làm mới được dữ liệu ({error.message}), đang hiển thị bản đã tải.
    </p>
  )
}
