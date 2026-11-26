import { X, type LucideIcon } from 'lucide-react'
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

interface DialogHeaderProps {
  readonly icon: LucideIcon
  readonly title: string
  readonly subtitle: string
}

export function DialogHeader({ icon: Icon, title, subtitle }: DialogHeaderProps): ReactElement {
  return (
    <header className="flex items-center gap-3.5 border-b border-line py-5 pr-4 pl-6">
      <span className="grid size-10 shrink-0 place-items-center rounded-[10px] bg-accent/20">
        <Icon className="size-[18px]" aria-hidden="true" />
      </span>
      <span className="grid min-w-0 flex-1 gap-0.5">
        <h2 className="text-[17px] font-semibold">{title}</h2>
        <span className="truncate text-[13px] text-muted">{subtitle}</span>
      </span>
      <button
        type="button"
        className="grid size-8 shrink-0 place-items-center rounded-lg bg-transparent p-0 text-muted hover:bg-ink/7 hover:text-ink"
        aria-label="Close"
        onClick={(event) => {
          event.currentTarget.closest('dialog')?.close()
        }}
      >
        <X className="size-[18px]" aria-hidden="true" />
      </button>
    </header>
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
