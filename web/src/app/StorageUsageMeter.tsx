import type { ReactElement } from 'react'
import { DISPLAY_LOCALE } from './display-locale'
import { ProgressBar } from './ProgressBar'

const BYTES_PER_GIGABYTE = 1_000_000_000
const GIGABYTE_FORMAT = new Intl.NumberFormat(DISPLAY_LOCALE, { maximumFractionDigits: 1 })

interface StorageUsageMeterProps {
  readonly usedBytes: number
  readonly quotaBytes: number
}

export function StorageUsageMeter({ usedBytes, quotaBytes }: StorageUsageMeterProps): ReactElement {
  return (
    <span className="flex items-center gap-2.5">
      <ProgressBar className="h-1.5 w-24 rounded-full" completed={usedBytes} total={quotaBytes} />
      <span className="tabular-nums">
        {gigabytes(usedBytes)} / {gigabytes(quotaBytes)} GB
      </span>
    </span>
  )
}

function gigabytes(bytes: number): string {
  return GIGABYTE_FORMAT.format(bytes / BYTES_PER_GIGABYTE)
}
