import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type ReactElement } from 'react'
import { Trash2 } from 'lucide-react'
import { removePublishedCourse } from '../../api/publish-queries'
import { Button, ButtonSize, ButtonVariant } from '../../ui/Button'

export function RemovePublishedButton(): ReactElement {
  const queryClient = useQueryClient()
  const [isConfirming, setIsConfirming] = useState(false)
  const remove = useMutation({
    mutationFn: removePublishedCourse,
    onSettled: () => {
      setIsConfirming(false)
      return queryClient.invalidateQueries()
    },
  })
  return (
    <span className="ml-auto flex items-center gap-3">
      {remove.isError && (
        <span className="text-danger" role="alert">
          {remove.error.message}
        </span>
      )}
      <Button
        variant={ButtonVariant.Danger}
        size={ButtonSize.Small}
        icon={Trash2}
        disabled={remove.isPending}
        onClick={() => {
          if (isConfirming) remove.mutate()
          else setIsConfirming(true)
        }}
      >
        {isConfirming ? 'Remove from R2?' : 'Remove'}
      </Button>
    </span>
  )
}
