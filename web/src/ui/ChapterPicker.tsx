import type { ReactElement, ReactNode } from 'react'
import { toggled } from './toggled'

export interface PickableChapter {
  readonly dir: string
  readonly detail: ReactNode
}

interface ChapterPickerProps {
  readonly chapters: readonly PickableChapter[]
  readonly selected: ReadonlySet<string>
  readonly onChange: (selected: ReadonlySet<string>) => void
}

export function ChapterPicker({ chapters, selected, onChange }: ChapterPickerProps): ReactElement {
  const isAllSelected = selected.size === chapters.length
  return (
    <fieldset className="grid gap-2">
      <legend className="mb-2 flex w-full items-center justify-between text-[13px] font-semibold">
        Chapters
        <button
          type="button"
          className="bg-transparent px-0 py-0 text-xs font-semibold text-accent-soft hover:underline"
          onClick={() => {
            onChange(new Set(isAllSelected ? [] : chapters.map((chapter) => chapter.dir)))
          }}
        >
          {isAllSelected ? 'None' : 'All'}
        </button>
      </legend>
      <ul className="m-0 grid max-h-64 [scrollbar-width:thin] list-none gap-1 overflow-y-auto rounded-xl bg-page p-1.5">
        {chapters.map((chapter) => (
          <li key={chapter.dir}>
            <label className="flex cursor-pointer items-center gap-3 rounded-lg px-2.5 py-2 text-[13px] hover:bg-ink/5">
              <input
                type="checkbox"
                className="size-4 shrink-0 accent-accent"
                checked={selected.has(chapter.dir)}
                onChange={() => {
                  onChange(toggled(selected, chapter.dir))
                }}
              />
              <span className="min-w-0 flex-1 truncate">{chapter.dir}</span>
              <span className="flex shrink-0 items-center gap-2 text-xs text-muted tabular-nums">
                {chapter.detail}
              </span>
            </label>
          </li>
        ))}
      </ul>
    </fieldset>
  )
}
