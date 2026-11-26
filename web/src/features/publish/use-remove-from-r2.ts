import { useMutation, useQueryClient, type UseMutationResult } from '@tanstack/react-query'
import { removePublishedCourse } from '../../api/publish-queries'

export type RemoveFromR2 = UseMutationResult<void, Error, void>

export function useRemoveFromR2(onRemoved?: () => void): RemoveFromR2 {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: removePublishedCourse,
    onSuccess: onRemoved,
    onSettled: () => queryClient.invalidateQueries(),
  })
}
