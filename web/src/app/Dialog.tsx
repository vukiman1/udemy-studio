import { X } from 'lucide-react'
import { useEffect, useRef, type ReactElement, type ReactNode } from 'react'

interface DialogProps {
  readonly title: string
  readonly onClose: () => void
  readonly children: ReactNode
}

// A native modal dialog: the browser handles focus, Escape and the backdrop.
export function Dialog({ title, onClose, children }: DialogProps): ReactElement {
  const dialogRef = useRef<HTMLDialogElement | null>(null)
  useEffect(() => {
    dialogRef.current?.showModal()
  }, [])

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      className="m-auto w-[min(560px,92vw)] rounded-xl border border-line bg-surface p-0 text-ink backdrop:bg-black/50"
    >
      <header className="flex items-center gap-3 border-b border-line px-5 py-4">
        <h2 className="flex-1 font-semibold wrap-anywhere">{title}</h2>
        <button
          type="button"
          className="link-button text-muted"
          aria-label="Đóng"
          onClick={() => {
            dialogRef.current?.close()
          }}
        >
          <X className="size-5" />
        </button>
      </header>
      <div className="grid gap-4 px-5 py-4">{children}</div>
    </dialog>
  )
}
