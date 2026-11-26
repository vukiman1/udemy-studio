import type { ReactElement } from 'react'

interface StatTileProps {
  readonly label: string
  readonly value: string
  readonly note?: string
}

export function StatTile({ label, value, note }: StatTileProps): ReactElement {
  return (
    <div className="grid min-w-0 flex-1 basis-32 content-start gap-1 rounded-xl bg-page px-4 py-3.5">
      <span className="text-xs text-muted">{label}</span>
      <span className="text-xl font-semibold">{value}</span>
      {note !== undefined && <span className="text-xs text-muted">{note}</span>}
    </div>
  )
}
