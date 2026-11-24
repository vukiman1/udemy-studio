import { useMutation } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { listAIModels } from '../../api/ai-queries'
import type { AIModel } from '../../api/types'
import type { AIDraftEditor } from './use-ai-draft'

interface ModelPickerProps {
  readonly editor: AIDraftEditor
}

const ADD_PROMPT_VALUE = ''

export function ModelPicker({ editor }: ModelPickerProps): ReactElement {
  const { draft, provider } = editor
  const listing = useMutation({ mutationFn: () => listAIModels(draft.provider, draft.apiKey) })
  const offered = mergeModels(provider.models, listing.data ?? [])
  const labelOf = (id: string): string => offered.find((model) => model.id === id)?.label ?? id
  const addable = offered.filter((model) => !draft.models.includes(model.id))

  return (
    <div className="grid gap-2">
      <span className="font-semibold">Model</span>
      <div className="flex flex-wrap items-center gap-2 rounded-lg border border-line bg-page px-2 py-2">
        {draft.models.map((model) => (
          <span
            key={model}
            className="inline-flex items-center gap-1 rounded-md bg-surface-strong px-2 py-1"
          >
            {labelOf(model)}
            <button
              type="button"
              aria-label={`Bỏ ${labelOf(model)}`}
              className="link-button text-muted"
              onClick={() => {
                editor.chooseModels(draft.models.filter((chosen) => chosen !== model))
              }}
            >
              ×
            </button>
          </span>
        ))}
        <select
          className="min-w-40 flex-1 border-0 bg-transparent px-1 py-1 text-muted"
          aria-label="Thêm model"
          value={ADD_PROMPT_VALUE}
          onChange={(event) => {
            if (event.target.value !== ADD_PROMPT_VALUE)
              editor.chooseModels([...draft.models, event.target.value])
          }}
        >
          <option value={ADD_PROMPT_VALUE}>Thêm model…</option>
          {addable.map((model) => (
            <option key={model.id} value={model.id}>
              {model.label}
            </option>
          ))}
        </select>
      </div>
      <p className="m-0 text-[0.9rem] text-muted">
        Model đầu tiên được gọi trước; các model sau là dự phòng khi model trước lỗi.
      </p>
      <div className="flex flex-wrap gap-4">
        <button
          type="button"
          className="link-button"
          disabled={listing.isPending}
          onClick={() => {
            listing.mutate()
          }}
        >
          {listing.isPending ? 'Đang tải danh sách…' : 'Xem tất cả model'}
        </button>
      </div>
      {listing.isSuccess && (
        <p className="m-0 text-[0.9rem] text-muted">
          {provider.label} có {listing.data.length} model; chọn thêm trong ô &quot;Thêm
          model…&quot;.
        </p>
      )}
      {listing.isError && (
        <p className="m-0 text-[0.9rem] text-danger" role="alert">
          {listing.error.message}
        </p>
      )}
    </div>
  )
}

function mergeModels(known: readonly AIModel[], listed: readonly AIModel[]): readonly AIModel[] {
  const knownIds = new Set(known.map((model) => model.id))
  return [...known, ...listed.filter((model) => !knownIds.has(model.id))]
}
