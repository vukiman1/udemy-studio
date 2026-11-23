import type { ReactElement } from 'react'
import { articleUrl } from '../../api/queries'

interface ArticleFrameProps {
  readonly lectureId: string
  readonly title: string
}

export function ArticleFrame({ lectureId, title }: ArticleFrameProps): ReactElement {
  return (
    <iframe
      className="h-[70dvh] w-full rounded-lg border border-line bg-white"
      src={articleUrl(lectureId)}
      title={title}
      sandbox=""
      referrerPolicy="no-referrer"
    />
  )
}
