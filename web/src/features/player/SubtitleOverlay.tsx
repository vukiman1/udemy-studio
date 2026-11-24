import type { CSSProperties, ReactElement } from 'react'
import type { CueLine } from '../../api/types'

interface SubtitleOverlayProps {
  readonly lines: readonly CueLine[]
  readonly fontScale: number
  readonly textOpacity: number
  readonly backgroundOpacity: number
}

type OverlayStyle = CSSProperties &
  Record<'--font-scale' | '--subtitle-text-opacity' | '--subtitle-background-opacity', string>

// The shadow keeps the text readable once the background is made see-through.
const LINE =
  'rounded-[0.25em] bg-[rgb(0_0_0/var(--subtitle-background-opacity))] px-[0.45em] py-[0.1em] leading-[1.35] [text-shadow:0_1px_2px_rgb(0_0_0/0.8)]'
const PRIMARY_LINE = `${LINE} text-[rgb(255_255_255/var(--subtitle-text-opacity))]`
const SECONDARY_LINE = `${LINE} text-[0.82em] text-[rgb(230_230_230/var(--subtitle-text-opacity))]`

export function SubtitleOverlay({
  lines,
  fontScale,
  textOpacity,
  backgroundOpacity,
}: SubtitleOverlayProps): ReactElement | null {
  if (lines.length === 0) return null
  const style: OverlayStyle = {
    '--font-scale': String(fontScale),
    '--subtitle-text-opacity': String(textOpacity),
    '--subtitle-background-opacity': String(backgroundOpacity),
  }
  return (
    <div
      className="pointer-events-none absolute bottom-[var(--subtitle-offset,8%)] left-1/2 grid w-[min(92%,60em)] -translate-x-1/2 justify-items-center gap-[0.2em] text-center text-[calc(clamp(14px,2.4cqw,30px)*var(--font-scale))] transition-[bottom] duration-250"
      style={style}
      aria-live="off"
    >
      {lines.map((line, position) => (
        <p
          key={`${line.language}-${String(position)}`}
          lang={line.language}
          className={position === 0 ? PRIMARY_LINE : SECONDARY_LINE}
        >
          {line.text}
        </p>
      ))}
    </div>
  )
}
