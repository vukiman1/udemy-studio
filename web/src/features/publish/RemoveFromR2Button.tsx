import { useState, type ReactElement } from 'react'
import { Trash2 } from 'lucide-react'
import { Button, ButtonSize, ButtonVariant } from '../../ui/Button'
import type { RemoveFromR2 } from './use-remove-from-r2'

interface RemoveFromR2ButtonProps {
  readonly removal: RemoveFromR2
  readonly label: string
  readonly size?: ButtonSize
}

export function RemoveFromR2Button({
  removal,
  label,
  size,
}: RemoveFromR2ButtonProps): ReactElement {
  const [isConfirming, setIsConfirming] = useState(false)
  return (
    <Button
      variant={ButtonVariant.Danger}
      size={size}
      icon={Trash2}
      disabled={removal.isPending}
      onClick={() => {
        if (!isConfirming) {
          setIsConfirming(true)
          return
        }
        setIsConfirming(false)
        removal.mutate()
      }}
    >
      {isConfirming ? 'Remove from R2?' : label}
    </Button>
  )
}
