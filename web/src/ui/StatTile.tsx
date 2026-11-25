import type { ReactElement } from 'react'

interface StatTileProps {
  readonly label: string
  readonly value: string
}

export function StatTile({ label, value }: StatTileProps): ReactElement {
  return (
    <div className="grid min-w-0 flex-1 basis-32 gap-1 rounded-xl bg-page px-4 py-3.5">
      <span className="text-xs text-muted">{label}</span>
      <span className="text-xl font-semibold">{value}</span>
    </div>
  )
}
