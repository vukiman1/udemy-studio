import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { useAIProviders, useAISettings } from '../../api/ai-queries'
import { QUERY_KEYS } from '../../api/queries'
import { cancelTranslation, isTranslating, startTranslation } from '../../api/translation-queries'
import type { TranslationRun } from '../../api/types'
import { Dialog } from '../../app/Dialog'
import { AI_SETTINGS_PATH } from '../../app/paths'
import { RunInBackgroundButton } from '../../app/RunInBackgroundButton'
import { TranslationPlanSummary } from './TranslationPlanSummary'
import { TranslationRunStatus } from './TranslationRunStatus'

interface TranslationDialogProps {
  readonly courseId: string
  readonly title: string
  readonly run: TranslationRun | undefined
  readonly onClose: () => void
}

export function TranslationDialog({
  courseId,
  title,
  run,
  onClose,
}: TranslationDialogProps): ReactElement {
  const queryClient = useQueryClient()
  const settings = useAISettings()
  const providers = useAIProviders()
  const refresh = (): Promise<void> =>
    queryClient.invalidateQueries({ queryKey: QUERY_KEYS.courseTools })
  const start = useMutation({ mutationFn: () => startTranslation(courseId), onSuccess: refresh })
  const cancel = useMutation({ mutationFn: () => cancelTranslation(courseId), onSuccess: refresh })
  const activeId = settings.data?.provider
  const provider = providers.data?.find((candidate) => candidate.id === activeId)
  const isReady = settings.data?.configured === true && settings.data.enabled
  const error = start.error ?? cancel.error

  return (
    <Dialog title={`Dịch phụ đề · ${title}`} onClose={onClose}>
      <p className="m-0 text-[0.9rem] text-muted">
        {isReady
          ? `Đang dùng ${provider?.label ?? ''} · ${provider?.model.label ?? ''}. `
          : 'Chưa bật AI nào. '}
        <Link to={AI_SETTINGS_PATH} className="text-accent underline">
          Cài đặt AI
        </Link>
      </p>
      {run !== undefined && <TranslationRunStatus run={run} />}
      {run !== undefined && isTranslating(run) ? (
        <div className="flex gap-3">
          <RunInBackgroundButton onClick={onClose} />
          <button
            type="button"
            disabled={cancel.isPending}
            onClick={() => {
              cancel.mutate()
            }}
          >
            Dừng dịch
          </button>
        </div>
      ) : (
        isReady && (
          <TranslationPlanSummary
            courseId={courseId}
            isStarting={start.isPending}
            onStart={() => {
              start.mutate()
            }}
            onDismiss={onClose}
          />
        )
      )}
      {error !== null && (
        <p className="m-0 text-[0.9rem] text-danger" role="alert">
          {error.message}
        </p>
      )}
    </Dialog>
  )
}
