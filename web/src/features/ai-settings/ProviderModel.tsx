import type { ReactElement } from 'react'
import type { AIProvider } from '../../api/types'
import { SETTINGS_LABEL, SETTINGS_ROW } from '../../ui/settings-row'
import { DISPLAY_LOCALE } from '../../app/display-locale'

const PRICE_FORMAT = new Intl.NumberFormat(DISPLAY_LOCALE, {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 2,
})

interface ProviderModelProps {
  readonly provider: AIProvider
}

// The model is fixed per provider, chosen by measuring translations, so a run never moves onto a pricier one.
export function ProviderModel({ provider }: ProviderModelProps): ReactElement {
  return (
    <div className={SETTINGS_ROW}>
      <span className={SETTINGS_LABEL}>Model</span>
      <span>
        {provider.model.label}
        {` · ${PRICE_FORMAT.format(provider.inputUsdPerMillion)} in, ${PRICE_FORMAT.format(provider.outputUsdPerMillion)} out per million tokens`}
      </span>
    </div>
  )
}
