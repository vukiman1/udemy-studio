import type { LucideIcon } from 'lucide-react'
import type { ButtonHTMLAttributes, ReactElement } from 'react'
import { Link, type LinkProps } from 'react-router'

export enum ButtonVariant {
  Primary = 'primary',
  Secondary = 'secondary',
  Ghost = 'ghost',
  Danger = 'danger',
}

export enum ButtonSize {
  Small = 'small',
  Medium = 'medium',
  Large = 'large',
}

const VARIANT_CLASSES: Readonly<Record<ButtonVariant, string>> = {
  [ButtonVariant.Primary]: 'bg-accent text-on-accent not-disabled:hover:brightness-110',
  [ButtonVariant.Secondary]: 'bg-button text-ink not-disabled:hover:brightness-125',
  [ButtonVariant.Ghost]: 'bg-transparent text-ink not-disabled:hover:bg-ink/7',
  [ButtonVariant.Danger]: 'bg-danger-soft text-danger not-disabled:hover:brightness-110',
}

const SIZE_CLASSES: Readonly<Record<ButtonSize, string>> = {
  [ButtonSize.Small]: 'h-8 gap-1.5 rounded-lg px-3 text-xs [&_svg]:size-3.5',
  [ButtonSize.Medium]: 'h-[42px] gap-[7px] rounded-[10px] px-4 text-[13px] [&_svg]:size-[18px]',
  [ButtonSize.Large]: 'h-12 gap-2 rounded-xl px-5 text-[15px] [&_svg]:size-[18px]',
}

function buttonClass(variant: ButtonVariant, size: ButtonSize): string {
  return `inline-flex shrink-0 items-center justify-center font-semibold whitespace-nowrap transition-[translate,scale,background-color,filter] duration-200 ease-soft motion-safe:not-disabled:hover:-translate-y-0.5 motion-safe:not-disabled:active:translate-y-0 motion-safe:not-disabled:active:scale-[0.97] ${VARIANT_CLASSES[variant]} ${SIZE_CLASSES[size]}`
}

interface ButtonLook {
  readonly variant?: ButtonVariant
  readonly size?: ButtonSize
  readonly icon?: LucideIcon
}

type ActionButtonProps = ButtonLook & ButtonHTMLAttributes<HTMLButtonElement>

type LinkButtonProps = ButtonLook & LinkProps

type ButtonProps = ActionButtonProps | LinkButtonProps

export function Button({
  variant = ButtonVariant.Secondary,
  size = ButtonSize.Medium,
  icon: Icon,
  className = '',
  children,
  ...target
}: ButtonProps): ReactElement {
  const classes = `${buttonClass(variant, size)} ${className}`
  const content = (
    <>
      {Icon !== undefined && <Icon aria-hidden="true" />}
      {children}
    </>
  )
  if ('to' in target)
    return (
      <Link className={classes} {...target}>
        {content}
      </Link>
    )
  const { type = 'button', ...attributes } = target
  return (
    <button type={type} className={classes} {...attributes}>
      {content}
    </button>
  )
}
