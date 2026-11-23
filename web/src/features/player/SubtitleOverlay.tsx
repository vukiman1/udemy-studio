import type { CSSProperties, ReactElement } from 'react'
import type { CueLine } from '../../api/types'
import styles from './SubtitleOverlay.module.css'

interface SubtitleOverlayProps {
  readonly lines: readonly CueLine[]
  readonly fontScale: number
}

type OverlayStyle = CSSProperties & Record<'--font-scale', string>

export function SubtitleOverlay({ lines, fontScale }: SubtitleOverlayProps): ReactElement | null {
  if (lines.length === 0) return null
  const style: OverlayStyle = { '--font-scale': String(fontScale) }
  return (
    <div className={styles.overlay} style={style} aria-live="off">
      {lines.map((line, position) => (
        <p
          key={`${line.language}-${String(position)}`}
          lang={line.language}
          className={position === 0 ? styles.primary : styles.secondary}
        >
          {line.text}
        </p>
      ))}
    </div>
  )
}
