import type { ReactElement } from 'react'
import { articleUrl } from '../../api/queries'

interface ArticleFrameProps {
  readonly lectureId: string
  readonly title: string
  readonly canRunScripts: boolean
}

// A quiz page needs its own scripts; without allow-same-origin they still run in an opaque origin, away from the app.
const QUIZ_SANDBOX = 'allow-scripts'

export function ArticleFrame({ lectureId, title, canRunScripts }: ArticleFrameProps): ReactElement {
  return (
    <iframe
      className="h-[70dvh] w-full rounded-2xl bg-card"
      src={articleUrl(lectureId)}
      title={title}
      sandbox={canRunScripts ? QUIZ_SANDBOX : ''}
      referrerPolicy="no-referrer"
    />
  )
}
