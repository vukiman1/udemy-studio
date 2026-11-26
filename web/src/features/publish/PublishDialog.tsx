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
import { RemoveFromR2Button } from './RemoveFromR2Button'
import { useRemoveFromR2 } from './use-remove-from-r2'

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
  const removal = useRemoveFromR2(onClose)
  const error = start.error ?? cancel.error ?? removal.error
  const errorNote = error !== null && (
    <p className="text-[13px] text-danger" role="alert">
      {error.message}
    </p>
  )

  return (
    <Dialog label={`Publish to R2: ${title}`} onClose={onClose}>
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
              disabled={cancel.isPending || removal.isPending}
              onClick={() => {
                cancel.mutate()
              }}
            >
              Stop
            </Button>
            <RemoveFromR2Button removal={removal} label="Stop & remove" />
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
          removeButton={<RemoveFromR2Button removal={removal} label="Remove from R2" />}
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
