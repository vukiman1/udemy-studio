import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { QUERY_KEYS, saveProgress } from '../../api/queries'
import type { LectureSummary } from '../../api/types'
import { CheckIcon } from '@videojs/react/icons'

interface MarkCompleteButtonProps {
  readonly lecture: LectureSummary
  readonly courseId: string
}

const READ_THROUGH = { positionSeconds: 1, durationSeconds: 1, completed: true } as const

const COMPLETION_TEXT = {
  article: { done: 'Đã đọc xong bài này.', action: 'Đánh dấu đã đọc' },
  quiz: { done: 'Đã làm xong bài này.', action: 'Đánh dấu đã làm' },
} as const

export function MarkCompleteButton({ lecture, courseId }: MarkCompleteButtonProps): ReactElement {
  const queryClient = useQueryClient()
  const text = COMPLETION_TEXT[lecture.kind === 'quiz' ? 'quiz' : 'article']
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

  if (lecture.progress?.completed === true) {
    return (
      <p className="inline-flex items-center gap-1 text-accent">
        <CheckIcon className="size-4.5" />
        {text.done}
      </p>
    )
  }
  return (
    <div>
      <button
        type="button"
        disabled={markComplete.isPending}
        onClick={() => {
          markComplete.mutate()
        }}
      >
        {text.action}
      </button>
      {markComplete.isError && <p role="alert">Không lưu được: {markComplete.error.message}</p>}
    </div>
  )
}
