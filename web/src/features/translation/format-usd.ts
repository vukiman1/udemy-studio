const SMALLEST_SHOWN_USD = 0.01
const USD_FORMAT = new Intl.NumberFormat('vi-VN', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 2,
})

// Cheap models cost fractions of a cent, which would otherwise round to a misleading 0.
export function formatUsd(amount: number): string {
  if (amount > 0 && amount < SMALLEST_SHOWN_USD)
    return `dưới ${USD_FORMAT.format(SMALLEST_SHOWN_USD)}`
  return USD_FORMAT.format(amount)
}
