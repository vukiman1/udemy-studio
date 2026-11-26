import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { Pause } from 'lucide-react'
import { cancelPublish, isPublishing, startPublish } from '../../api/publish-queries'
import { QUERY_KEYS } from '../../api/queries'
import type { PublishRun } from '../../api/types'
import { Dialog, DialogBody, DialogFooter } from '../../app/Dialog'
import { RunInBackgroundButton } from '../../app/RunInBackgroundButton'
import { Button, ButtonVariant } from '../../ui/Button'
import { PublishPlanSummary } from './PublishPlanSummary'
import { PublishProgress } from './PublishProgress'
import { PublishResult } from './PublishResult'

interface PublishDialogProps {
  readonly courseId: string
  readonly title: string
  readonly run: PublishRun | undefined
  readonly onClose: () => void
}

export function PublishDialog({ courseId, title, run, onClose }: PublishDialogProps): ReactElement {
  const queryClient = useQueryClient()
  const refresh = (): Promise<void> =>
    queryClient.invalidateQueries({ queryKey: QUERY_KEYS.publishRun })
  const start = useMutation({
    mutationFn: (chapters: readonly string[]) => startPublish(courseId, chapters),
    onSuccess: refresh,
  })
  const cancel = useMutation({ mutationFn: cancelPublish, onSuccess: refresh })
  const error = start.error ?? cancel.error
  const errorNote = error !== null && (
    <p className="text-[13px] text-danger" role="alert">
      {error.message}
    </p>
  )

  return (
    <Dialog label={`Đưa lên R2: ${title}`} onClose={onClose}>
      {run !== undefined && isPublishing(run) ? (
        <>
          <DialogBody>
            <PublishProgress run={run} />
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
              Dừng
            </Button>
            <RunInBackgroundButton onClick={onClose} />
          </DialogFooter>
        </>
      ) : (
        <PublishPlanSummary
          courseId={courseId}
          isStarting={start.isPending}
          onStart={(chapters) => {
            start.mutate(chapters)
          }}
          onDismiss={onClose}
        >
          {run !== undefined && run.state !== 'running' && (
            <PublishResult run={run} state={run.state} />
          )}
          {errorNote}
        </PublishPlanSummary>
      )}
    </Dialog>
  )
}
