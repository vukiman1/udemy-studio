import type { CSSProperties, ReactElement } from 'react'
import type { CueLine } from '../../api/types'

interface SubtitleOverlayProps {
  readonly lines: readonly CueLine[]
  readonly fontScale: number
}

type OverlayStyle = CSSProperties & Record<'--font-scale', string>

const PRIMARY_LINE = 'rounded-[0.25em] bg-black/72 px-[0.45em] py-[0.1em] leading-[1.35] text-white'
const SECONDARY_LINE =
  'rounded-[0.25em] bg-black/72 px-[0.45em] py-[0.1em] text-[0.82em] leading-[1.35] text-[#e6e6e6]'

export function SubtitleOverlay({ lines, fontScale }: SubtitleOverlayProps): ReactElement | null {
  if (lines.length === 0) return null
  const style: OverlayStyle = { '--font-scale': String(fontScale) }
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
