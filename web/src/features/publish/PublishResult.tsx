import type { ReactElement } from 'react'
import { CircleAlert, CircleCheck, CircleStop, type LucideIcon } from 'lucide-react'
import type { PublishRun } from '../../api/types'

type FinishedState = Exclude<PublishRun['state'], 'running'>

interface ResultLook {
  readonly icon: LucideIcon
  readonly box: string
  readonly accent: string
  readonly heading: string
}

const RESULT_LOOKS: Readonly<Record<FinishedState, ResultLook>> = {
  done: {
    icon: CircleCheck,
    box: 'bg-success-soft',
    accent: 'text-success',
    heading: 'Đã đưa lên R2',
  },
  cancelled: { icon: CircleStop, box: 'bg-page', accent: 'text-muted', heading: 'Đã dừng' },
  failed: {
    icon: CircleAlert,
    box: 'bg-danger-soft',
    accent: 'text-danger',
    heading: 'Đưa lên bị lỗi',
  },
}

interface PublishResultProps {
  readonly run: PublishRun
  readonly state: FinishedState
}

export function PublishResult({ run, state }: PublishResultProps): ReactElement {
  const look = RESULT_LOOKS[state]
  const Icon = look.icon
  return (
    <div className={`flex gap-3 rounded-xl px-4 py-3.5 text-[13px] ${look.box}`}>
      <Icon className={`mt-0.5 size-[18px] shrink-0 ${look.accent}`} aria-hidden="true" />
      <div className="grid gap-0.5">
        <p className={`text-[15px] font-semibold ${look.accent}`}>{look.heading}</p>
        <p className="wrap-anywhere">{detail(run, state)}</p>
      </div>
    </div>
  )
}

function detail(run: PublishRun, state: FinishedState): string {
  const counts = `${String(run.sentFiles)} file gửi mới, ${String(run.skippedFiles)} file đã có sẵn.`
  if (state === 'done') return counts
  if (state === 'cancelled')
    return `${counts} Bản trên VPS chưa thấy bản mới cho đến khi đưa lên xong.`
  return run.failure ?? 'Không rõ lỗi.'
}
