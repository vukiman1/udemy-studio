import type { ReactElement } from 'react'
import { articleUrl } from '../../api/queries'
import styles from './ArticleFrame.module.css'

interface ArticleFrameProps {
  readonly lectureId: string
  readonly title: string
}

export function ArticleFrame({ lectureId, title }: ArticleFrameProps): ReactElement {
  return (
    <iframe
      className={styles.frame}
      src={articleUrl(lectureId)}
      title={title}
      sandbox=""
      referrerPolicy="no-referrer"
    />
  )
}
