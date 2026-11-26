import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { QUERY_KEYS, saveProgress } from '../../api/queries'
import type { LectureSummary } from '../../api/types'
import { CircleCheck } from 'lucide-react'
import { Button, ButtonSize, ButtonVariant } from '../../ui/Button'

interface MarkCompleteButtonProps {
  readonly lecture: LectureSummary
  readonly courseId: string
}

const READ_THROUGH = { positionSeconds: 1, durationSeconds: 1, completed: true } as const

const COMPLETION_TEXT = {
  article: { done: 'Read.', action: 'Mark as read' },
  quiz: { done: 'Done.', action: 'Mark as done' },
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
      ])
    },
  })

  if (lecture.progress?.completed === true) {
    return (
      <p className="inline-flex items-center gap-1.5 text-[13px] font-medium text-accent-soft">
        <CircleCheck className="size-4" aria-hidden="true" />
        {text.done}
      </p>
    )
  }
  return (
    <div className="flex flex-wrap items-center gap-3">
      <Button
        variant={ButtonVariant.Ghost}
        size={ButtonSize.Small}
        icon={CircleCheck}
        disabled={markComplete.isPending}
        onClick={() => {
          markComplete.mutate()
        }}
      >
        {text.action}
      </Button>
      {markComplete.isError && (
        <p className="text-[13px] text-danger" role="alert">
          Could not save: {markComplete.error.message}
        </p>
      )}
    </div>
  )
}
