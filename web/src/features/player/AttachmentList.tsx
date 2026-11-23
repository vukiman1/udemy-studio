import type { ReactElement } from 'react'
import { attachmentUrl } from '../../api/queries'
import type { LectureSummary } from '../../api/types'

interface AttachmentListProps {
  readonly attachments: LectureSummary['attachments']
}

export function AttachmentList({ attachments }: AttachmentListProps): ReactElement | null {
  if (attachments.length === 0) return null
  return (
    <section>
      <h3 className="my-2 text-[0.95rem] font-semibold">Tài liệu kèm theo</h3>
      <ul className="grid list-disc gap-1 pl-5 wrap-anywhere">
        {attachments.map((attachment) => (
          <li key={attachment.id}>
            <a className="text-accent underline" href={attachmentUrl(attachment.id)} download>
              {attachment.name}
            </a>
          </li>
        ))}
      </ul>
    </section>
  )
}
