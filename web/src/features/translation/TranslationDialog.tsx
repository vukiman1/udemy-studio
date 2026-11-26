import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { Link } from 'react-router'
import { Info, Languages, Pause } from 'lucide-react'
import { useAIProviders, useAISettings } from '../../api/ai-queries'
import { QUERY_KEYS } from '../../api/queries'
import { cancelTranslation, isTranslating, startTranslation } from '../../api/translation-queries'
import type { AIProvider, TranslationRun } from '../../api/types'
import { Dialog, DialogBody, DialogFooter, DialogHeader } from '../../app/Dialog'
import { AI_SETTINGS_PATH } from '../../app/paths'
import { RunInBackgroundButton } from '../../app/RunInBackgroundButton'
import { Button, ButtonSize, ButtonVariant, buttonClass } from '../../ui/Button'
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
  const provider = providers.data?.find((candidate) => candidate.id === settings.data?.provider)
  const isReady = settings.data?.configured === true && settings.data.enabled
  const error = start.error ?? cancel.error
  const errorNote = error !== null && (
    <p className="text-[13px] text-danger" role="alert">
      {error.message}
    </p>
  )

  return (
    <Dialog label="Translate subtitles" onClose={onClose}>
      <DialogHeader icon={Languages} title="Translate subtitles" subtitle={title} />
      {run !== undefined && isTranslating(run) ? (
        <>
          <DialogBody>
            <ActiveAI provider={provider} />
            <TranslationRunStatus run={run} />
            {errorNote}
          </DialogBody>
          <DialogFooter>
            <Button
              variant={ButtonVariant.Ghost}
              icon={Pause}
              disabled={cancel.isPending}
              onClick={() => {
                cancel.mutate()
              }}
            >
              Stop
            </Button>
            <RunInBackgroundButton onClick={onClose} />
          </DialogFooter>
        </>
      ) : isReady ? (
        <TranslationPlanSummary
          courseId={courseId}
          isStarting={start.isPending}
          onStart={() => {
            start.mutate()
          }}
          onDismiss={onClose}
        >
          <ActiveAI provider={provider} />
          {run !== undefined && <TranslationRunStatus run={run} />}
          {errorNote}
        </TranslationPlanSummary>
      ) : (
        <>
          <DialogBody>
            {run !== undefined && <TranslationRunStatus run={run} />}
            {settings.data === undefined ? (
              <p className="text-muted" role={settings.isError ? 'alert' : 'status'}>
                {settings.isError
                  ? `Could not read AI settings (${settings.error.message}).`
                  : 'Reading AI settings…'}
              </p>
            ) : (
              <NoAIEnabled />
            )}
          </DialogBody>
          <DialogFooter>
            <Button onClick={onClose}>Cancel</Button>
            <Link
              to={AI_SETTINGS_PATH}
              className={buttonClass(ButtonVariant.Primary, ButtonSize.Medium)}
            >
              Open AI settings
            </Link>
          </DialogFooter>
        </>
      )}
    </Dialog>
  )
}

interface ActiveAIProps {
  readonly provider: AIProvider | undefined
}

function ActiveAI({ provider }: ActiveAIProps): ReactElement {
  return (
    <p className="flex items-center justify-between gap-3 text-[13px] text-muted">
      <span className="min-w-0 truncate">
        Using {provider?.label ?? '…'} · {provider?.model.label ?? '…'}
      </span>
      <Link to={AI_SETTINGS_PATH} className="shrink-0 font-medium text-accent-soft">
        AI settings
      </Link>
    </p>
  )
}

function NoAIEnabled(): ReactElement {
  return (
    <div className="grid justify-items-center gap-1.5 py-3 text-center">
      <Info className="mb-1 size-6 text-accent-soft" aria-hidden="true" />
      <p className="text-[15px] font-semibold">No AI enabled</p>
      <p className="text-[13px] text-muted">
        Enable a provider in AI settings to translate subtitles.
      </p>
    </div>
  )
}
