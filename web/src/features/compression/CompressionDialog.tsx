import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { Pause } from 'lucide-react'
import { cancelCompression, isCompressing, startCompression } from '../../api/compression-queries'
import { QUERY_KEYS } from '../../api/queries'
import type { CompressionRun } from '../../api/types'
import { Dialog, DialogBody, DialogFooter } from '../../app/Dialog'
import { RunInBackgroundButton } from '../../app/RunInBackgroundButton'
import { Button, ButtonVariant } from '../../ui/Button'
import { CompressionPlanSummary } from './CompressionPlanSummary'
import { CompressionProgress } from './CompressionProgress'
import { CompressionResult } from './CompressionResult'

interface CompressionDialogProps {
  readonly courseId: string
  readonly title: string
  readonly run: CompressionRun | undefined
  readonly onClose: () => void
}

export function CompressionDialog({
  courseId,
  title,
  run,
  onClose,
}: CompressionDialogProps): ReactElement {
  const queryClient = useQueryClient()
  const refresh = (): Promise<void> =>
    queryClient.invalidateQueries({ queryKey: QUERY_KEYS.courseTools })
  const start = useMutation({
    mutationFn: (chapters: readonly string[]) => startCompression(courseId, chapters),
    onSuccess: refresh,
  })
  const cancel = useMutation({ mutationFn: () => cancelCompression(courseId), onSuccess: refresh })
  const error = start.error ?? cancel.error
  const errorNote = error !== null && (
    <p className="text-[13px] text-danger" role="alert">
      {error.message}
    </p>
  )

  return (
    <Dialog label={`Compress videos: ${title}`} onClose={onClose}>
      {run !== undefined && isCompressing(run) ? (
        <>
          <DialogBody>
            <CompressionProgress run={run} />
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
      ) : (
        <CompressionPlanSummary
          courseId={courseId}
          isStarting={start.isPending}
          onStart={(chapters) => {
            start.mutate(chapters)
          }}
          onDismiss={onClose}
        >
          {run !== undefined && <CompressionResult run={run} />}
          {errorNote}
        </CompressionPlanSummary>
      )}
    </Dialog>
  )
}
