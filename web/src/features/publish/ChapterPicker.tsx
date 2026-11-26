import type { ReactElement } from 'react'
import { Cloud } from 'lucide-react'
import type { PublishChapter } from '../../api/types'
import { formatBytes } from '../compression/format'
import { toggled } from './chapter-selection'

interface ChapterPickerProps {
  readonly chapters: readonly PublishChapter[]
  readonly selected: ReadonlySet<string>
  readonly onChange: (selected: ReadonlySet<string>) => void
}

export function ChapterPicker({ chapters, selected, onChange }: ChapterPickerProps): ReactElement {
  const isAllSelected = selected.size === chapters.length
  return (
    <fieldset className="grid gap-2">
      <legend className="mb-2 flex w-full items-center justify-between text-[13px] font-semibold">
        Chương
        <button
          type="button"
          className="bg-transparent px-0 py-0 text-xs font-semibold text-accent-soft hover:underline"
          onClick={() => {
            onChange(new Set(isAllSelected ? [] : chapters.map((chapter) => chapter.dir)))
          }}
        >
          {isAllSelected ? 'Bỏ hết' : 'Tất cả'}
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
              {chapter.published && (
                <Cloud className="size-3.5 shrink-0 text-accent-soft" aria-label="Đang trên R2" />
              )}
              <span className="shrink-0 text-xs text-muted tabular-nums">
                {formatBytes(chapter.sizeBytes)}
              </span>
            </label>
          </li>
        ))}
      </ul>
    </fieldset>
  )
}
