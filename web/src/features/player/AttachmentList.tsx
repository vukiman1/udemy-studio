import type { ReactElement } from 'react'
import { FileText } from 'lucide-react'
import { attachmentUrl } from '../../api/queries'
import type { LectureSummary } from '../../api/types'

interface AttachmentListProps {
  readonly attachments: LectureSummary['attachments']
}

export function AttachmentList({ attachments }: AttachmentListProps): ReactElement | null {
  if (attachments.length === 0) return null
  return (
    <section className="grid gap-2 rounded-2xl bg-card p-4" aria-labelledby="attachments-title">
      <h3 id="attachments-title" className="text-sm font-semibold">
        Tài liệu kèm theo
      </h3>
      <ul className="grid list-none text-[13px] text-accent-soft">
        {attachments.map((attachment) => (
          <li key={attachment.id}>
            <a
              className="flex items-center gap-2 rounded-lg px-2 py-1.5 wrap-anywhere hover:bg-ink/5"
              href={attachmentUrl(attachment.id)}
              download
            >
              <FileText className="size-4 shrink-0" aria-hidden="true" />
              {attachment.name}
            </a>
          </li>
        ))}
      </ul>
    </section>
  )
}
