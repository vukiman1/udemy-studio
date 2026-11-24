export const MS_PER_SECOND = 1000
const SECONDS_PER_MINUTE = 60
const MINUTES_PER_HOUR = 60
const CLOCK_FIELD_DIGITS = 2

function clockField(value: number): string {
  return String(value).padStart(CLOCK_FIELD_DIGITS, '0')
}

export function formatTimestamp(milliseconds: number): string {
  const totalSeconds = Math.floor(milliseconds / MS_PER_SECOND)
  const totalMinutes = Math.floor(totalSeconds / SECONDS_PER_MINUTE)
  const hours = Math.floor(totalMinutes / MINUTES_PER_HOUR)
  const clock = `${clockField(totalMinutes % MINUTES_PER_HOUR)}:${clockField(totalSeconds % SECONDS_PER_MINUTE)}`
  return hours > 0 ? `${String(hours)}:${clock}` : clock
}
