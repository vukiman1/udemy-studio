import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type ReactElement } from 'react'
import {
  cancelCompression,
  isCompressing,
  startCompression,
  useCompressionRun,
} from '../../api/compression-queries'
import { QUERY_KEYS } from '../../api/queries'
import { CompressionPlanSummary } from './CompressionPlanSummary'
import { CompressionProgress } from './CompressionProgress'
import { CompressionResult } from './CompressionResult'

interface CompressionCardProps {
  readonly courseId: string
  readonly title: string
}

export function CompressionCard({ courseId, title }: CompressionCardProps): ReactElement {
  const queryClient = useQueryClient()
  const run = useCompressionRun(courseId)
  const [isPlanning, setIsPlanning] = useState(false)
  const refreshRun = (): Promise<void> =>
    queryClient.invalidateQueries({ queryKey: QUERY_KEYS.compression(courseId) })
  const start = useMutation({
    mutationFn: () => startCompression(courseId),
    onSuccess: async () => {
      setIsPlanning(false)
      await refreshRun()
    },
  })
  const cancel = useMutation({
    mutationFn: () => cancelCompression(courseId),
    onSuccess: refreshRun,
  })
  const error = run.error ?? start.error ?? cancel.error

  return (
    <article className="grid gap-2 rounded-xl border border-line p-4">
      <h3 className="font-semibold wrap-anywhere">{title}</h3>
      {isCompressing(run.data) ? (
        <CompressionProgress
          run={run.data}
          isCancelling={cancel.isPending}
          onCancel={() => {
            cancel.mutate()
          }}
        />
      ) : (
        <>
          {run.data != null && <CompressionResult run={run.data} />}
          {isPlanning ? (
            <CompressionPlanSummary
              courseId={courseId}
              isStarting={start.isPending}
              onStart={() => {
                start.mutate()
              }}
              onDismiss={() => {
                setIsPlanning(false)
              }}
            />
          ) : (
            <div>
              <button
                type="button"
                disabled={run.isError}
                onClick={() => {
                  setIsPlanning(true)
                }}
              >
                Nén video…
              </button>
            </div>
          )}
        </>
      )}
      {error !== null && (
        <p className="m-0 text-[0.9rem] text-danger" role="alert">
          {error.message}
        </p>
      )}
    </article>
  )
}
