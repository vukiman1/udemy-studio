import { useEffect, useRef, type ReactElement, type ReactNode } from 'react'

interface DialogProps {
  readonly label: string
  readonly onClose: () => void
  readonly children: ReactNode
}

export function Dialog({ label, onClose, children }: DialogProps): ReactElement {
  const dialogRef = useRef<HTMLDialogElement | null>(null)
  useEffect(() => {
    dialogRef.current?.showModal()
  }, [])

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      aria-label={label}
      className="m-auto w-[min(560px,92vw)] overflow-hidden rounded-[20px] bg-card p-0 text-ink shadow-[0_30px_106px_rgba(0,0,0,0.6)] backdrop:bg-black/50 motion-safe:animate-dialog-in motion-safe:backdrop:animate-fade-in max-md:mb-0 max-md:w-full max-md:max-w-none max-md:rounded-b-none max-md:motion-safe:animate-sheet-up"
    >
      {children}
    </dialog>
  )
}

interface DialogPartProps {
  readonly children: ReactNode
}

export function DialogBody({ children }: DialogPartProps): ReactElement {
  return <div className="grid gap-4 px-6 py-5 text-sm">{children}</div>
}

export function DialogFooter({ children }: DialogPartProps): ReactElement {
  return (
    <div className="flex flex-wrap items-center justify-end gap-2.5 border-t border-line bg-page px-6 py-4">
      {children}
    </div>
  )
}
