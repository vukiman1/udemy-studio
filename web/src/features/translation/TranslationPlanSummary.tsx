import type { ReactElement, ReactNode } from 'react'
import { Languages } from 'lucide-react'
import { useTranslationPlan } from '../../api/translation-queries'
import { DialogBody, DialogFooter } from '../../app/Dialog'
import { Button, ButtonVariant } from '../../ui/Button'
import { StatTile } from '../../ui/StatTile'
import { formatUsd } from './format-usd'

interface TranslationPlanSummaryProps {
  readonly courseId: string
  readonly isStarting: boolean
  readonly onStart: () => void
  readonly onDismiss: () => void
  readonly children: ReactNode
}

export function TranslationPlanSummary({
  courseId,
  isStarting,
  onStart,
  onDismiss,
  children,
}: TranslationPlanSummaryProps): ReactElement {
  const plan = useTranslationPlan(courseId)
  const lectures = plan.data?.lectures ?? 0

  return (
    <>
      <DialogBody>
        {children}
        {plan.isError && (
          <p className="text-danger" role="alert">
            Could not plan the translation ({plan.error.message}).
          </p>
        )}
        {plan.isPending && <p className="text-muted">Reading subtitles…</p>}
        {plan.data !== undefined &&
          (lectures === 0 ? (
            <p>Nothing left to translate.</p>
          ) : (
            <div className="flex flex-wrap gap-3">
              <StatTile label="To translate" value={String(lectures)} />
              <StatTile label="Already translated" value={String(plan.data.alreadyTranslated)} />
              <StatTile label="Estimated cost" value={`≈ ${formatUsd(plan.data.estimatedUsd)}`} />
            </div>
          ))}
      </DialogBody>
      <DialogFooter>
        {lectures > 0 ? (
          <>
            <Button onClick={onDismiss}>Cancel</Button>
            <Button
              variant={ButtonVariant.Primary}
              icon={Languages}
              disabled={isStarting}
              onClick={onStart}
            >
              Translate
            </Button>
          </>
        ) : (
          <Button variant={ButtonVariant.Primary} onClick={onDismiss}>
            Close
          </Button>
        )}
      </DialogFooter>
    </>
  )
}
