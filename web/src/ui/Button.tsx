import type { LucideIcon } from 'lucide-react'
import type { ButtonHTMLAttributes, ReactElement } from 'react'

export enum ButtonVariant {
  Primary = 'primary',
  Secondary = 'secondary',
  Ghost = 'ghost',
  Danger = 'danger',
  Done = 'done',
}

export enum ButtonSize {
  Small = 'small',
  Medium = 'medium',
  Large = 'large',
}

const VARIANT_CLASSES: Readonly<Record<ButtonVariant, string>> = {
  [ButtonVariant.Primary]: 'bg-accent text-on-accent enabled:hover:brightness-110',
  [ButtonVariant.Secondary]: 'bg-button text-ink enabled:hover:brightness-125',
  [ButtonVariant.Ghost]: 'bg-transparent text-ink enabled:hover:bg-ink/7',
  [ButtonVariant.Danger]: 'bg-danger-soft text-danger enabled:hover:brightness-110',
  [ButtonVariant.Done]: 'bg-success-soft text-ink disabled:opacity-100 [&_svg]:text-success',
}

const SIZE_CLASSES: Readonly<Record<ButtonSize, string>> = {
  [ButtonSize.Small]: 'h-8 gap-1.5 rounded-lg px-3 text-xs [&_svg]:size-3.5',
  [ButtonSize.Medium]: 'h-[42px] gap-[7px] rounded-[10px] px-4 text-[13px] [&_svg]:size-[18px]',
  [ButtonSize.Large]: 'h-12 gap-2 rounded-xl px-5 text-[15px] [&_svg]:size-[18px]',
}

export function buttonClass(variant: ButtonVariant, size: ButtonSize): string {
  return `inline-flex shrink-0 items-center justify-center font-semibold whitespace-nowrap ${VARIANT_CLASSES[variant]} ${SIZE_CLASSES[size]}`
}

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  readonly variant?: ButtonVariant
  readonly size?: ButtonSize
  readonly icon?: LucideIcon
}

export function Button({
  variant = ButtonVariant.Secondary,
  size = ButtonSize.Medium,
  icon: Icon,
  className = '',
  type = 'button',
  children,
  ...attributes
}: ButtonProps): ReactElement {
  return (
    <button type={type} className={`${buttonClass(variant, size)} ${className}`} {...attributes}>
      {Icon !== undefined && <Icon aria-hidden="true" />}
      {children}
    </button>
  )
}
