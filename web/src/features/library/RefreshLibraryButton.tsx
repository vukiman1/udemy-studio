import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { RotateCw } from 'lucide-react'
import { refreshLibrary } from '../../api/server-queries'
import { Button } from '../../ui/Button'

const SPINNING_ICON_CLASS = '[&_svg]:animate-spin'

export function RefreshLibraryButton(): ReactElement {
  const queryClient = useQueryClient()
  const refresh = useMutation({
    mutationFn: refreshLibrary,
    onSuccess: () => queryClient.invalidateQueries(),
  })
  return (
    <div className="flex items-center gap-3">
      {refresh.isError && (
        <span className="text-[13px] text-danger" role="alert">
          {refresh.error.message}
        </span>
      )}
      <Button
        icon={RotateCw}
        className={refresh.isPending ? SPINNING_ICON_CLASS : ''}
        disabled={refresh.isPending}
        onClick={() => {
          refresh.mutate()
        }}
      >
        Làm mới
      </Button>
    </div>
  )
}
