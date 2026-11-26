import { DISPLAY_LOCALE } from '../../app/display-locale'

const BYTES_PER_MEGABYTE = 1_000_000
const BYTES_PER_GIGABYTE = 1_000_000_000
const SECONDS_PER_MINUTE = 60
const MINUTES_PER_HOUR = 60
const SIZE_FORMAT = new Intl.NumberFormat(DISPLAY_LOCALE, { maximumFractionDigits: 1 })

export function formatBytes(bytes: number): string {
  if (bytes >= BYTES_PER_GIGABYTE) return `${SIZE_FORMAT.format(bytes / BYTES_PER_GIGABYTE)} GB`
  return `${SIZE_FORMAT.format(bytes / BYTES_PER_MEGABYTE)} MB`
}

export function formatRoughDuration(seconds: number): string {
  const totalMinutes = Math.max(1, Math.round(seconds / SECONDS_PER_MINUTE))
  const hours = Math.floor(totalMinutes / MINUTES_PER_HOUR)
  const minutes = totalMinutes % MINUTES_PER_HOUR
  if (hours === 0) return `${String(minutes)} min`
  return minutes === 0 ? `${String(hours)} h` : `${String(hours)} h ${String(minutes)} min`
}
