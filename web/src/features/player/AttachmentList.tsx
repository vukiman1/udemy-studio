import type { ReactElement } from 'react'
import { attachmentUrl } from '../../api/queries'
import type { LectureSummary } from '../../api/types'
import styles from './AttachmentList.module.css'

interface AttachmentListProps {
  readonly attachments: LectureSummary['attachments']
}

export function AttachmentList({ attachments }: AttachmentListProps): ReactElement | null {
  if (attachments.length === 0) return null
  return (
    <section>
      <h3 className={styles.heading}>Tài liệu kèm theo</h3>
      <ul className={styles.list}>
        {attachments.map((attachment) => (
          <li key={attachment.id}>
            <a href={attachmentUrl(attachment.id)} download>
              {attachment.name}
            </a>
          </li>
        ))}
      </ul>
    </section>
  )
}
