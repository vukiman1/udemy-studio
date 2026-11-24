import type { ReactElement, ReactNode } from 'react'

interface SettingsCardProps {
  readonly title: string
  readonly children: ReactNode
}

export function SettingsCard({ title, children }: SettingsCardProps): ReactElement {
  return (
    <section className="rounded-xl border border-line bg-surface">
      <h2 className="border-b border-line px-6 py-4 text-lg font-semibold">{title}</h2>
      <div className="grid gap-6 px-6 py-5">{children}</div>
    </section>
  )
}
