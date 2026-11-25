import type { ReactElement, ReactNode } from 'react'

interface SettingsCardProps {
  readonly title: string
  readonly children: ReactNode
}

// The design shows no card heading, so the title is kept for screen readers only.
export function SettingsCard({ title, children }: SettingsCardProps): ReactElement {
  return (
    <section
      className="grid gap-5 rounded-[20px] bg-card px-5 pt-7 pb-6 md:px-[30px]"
      aria-labelledby="settings-card-title"
    >
      <h2 id="settings-card-title" className="sr-only">
        {title}
      </h2>
      {children}
    </section>
  )
}
