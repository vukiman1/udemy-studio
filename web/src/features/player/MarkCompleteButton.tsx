import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { QUERY_KEYS, saveProgress } from '../../api/queries'
import type { LectureSummary } from '../../api/types'

interface MarkCompleteButtonProps {
  readonly lecture: LectureSummary
  readonly courseId: string
}

const READ_THROUGH = { positionSeconds: 1, durationSeconds: 1, completed: true } as const

export function MarkCompleteButton({ lecture, courseId }: MarkCompleteButtonProps): ReactElement {
  const queryClient = useQueryClient()
  const markComplete = useMutation({
    mutationFn: () => saveProgress(lecture.id, READ_THROUGH, false),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: QUERY_KEYS.course(courseId) }),
        queryClient.invalidateQueries({ queryKey: QUERY_KEYS.courses }),
        queryClient.invalidateQueries({ queryKey: QUERY_KEYS.continueTarget }),
      ])
    },
  })

  if (lecture.progress?.completed === true) return <p>✓ Đã đọc xong bài này.</p>
  return (
    <div>
      <button
        type="button"
        disabled={markComplete.isPending}
        onClick={() => {
          markComplete.mutate()
        }}
      >
        Đánh dấu đã đọc
      </button>
      {markComplete.isError && <p role="alert">Không lưu được: {markComplete.error.message}</p>}
    </div>
  )
}
