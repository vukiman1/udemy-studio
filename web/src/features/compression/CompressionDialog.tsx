import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { cancelCompression, isCompressing, startCompression } from '../../api/compression-queries'
import { QUERY_KEYS } from '../../api/queries'
import type { CompressionRun } from '../../api/types'
import { Dialog } from '../../app/Dialog'
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
  const start = useMutation({ mutationFn: () => startCompression(courseId), onSuccess: refresh })
  const cancel = useMutation({ mutationFn: () => cancelCompression(courseId), onSuccess: refresh })
  const error = start.error ?? cancel.error

  return (
    <Dialog title={`Nén video (H.265) · ${title}`} onClose={onClose}>
      {run !== undefined && isCompressing(run) ? (
        <CompressionProgress
          run={run}
          isCancelling={cancel.isPending}
          onCancel={() => {
            cancel.mutate()
          }}
          onRunInBackground={onClose}
        />
      ) : (
        <>
          {run !== undefined && <CompressionResult run={run} />}
          <CompressionPlanSummary
            courseId={courseId}
            isStarting={start.isPending}
            onStart={() => {
              start.mutate()
            }}
            onDismiss={onClose}
          />
        </>
      )}
      {error !== null && (
        <p className="m-0 text-[0.9rem] text-danger" role="alert">
          {error.message}
        </p>
      )}
    </Dialog>
  )
}
