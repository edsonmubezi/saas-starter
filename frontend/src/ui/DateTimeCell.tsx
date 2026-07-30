// components/DateTimeCell.tsx
import React from 'react'

type DateInput = string | number | Date | null | undefined
type Mode = 'utc' | 'local'

type DateTimeCellProps = {
  value?: DateInput
  mode?: Mode          // 'utc' (default) or 'local'
  showSeconds?: boolean
  fallback?: React.ReactNode
  className?: string
}

export default function DateTimeCell({
  value,
  mode = 'utc',
  showSeconds = true,
  fallback = '—',
  className,
}: DateTimeCellProps) {
  if (value == null) return <span className={className}>{fallback}</span>

  const d = value instanceof Date ? value : new Date(value)

  // Guard against invalid Date
  if (isNaN(d.getTime())) return <span className={className}>{fallback}</span>

  const pad = (n: number) => n.toString().padStart(2, '0')

  const Y = mode === 'utc' ? d.getUTCFullYear() : d.getFullYear()
  const M = mode === 'utc' ? d.getUTCMonth() + 1 : d.getMonth() + 1
  const D = mode === 'utc' ? d.getUTCDate() : d.getDate()
  const h = mode === 'utc' ? d.getUTCHours() : d.getHours()
  const m = mode === 'utc' ? d.getUTCMinutes() : d.getMinutes()
  const s = mode === 'utc' ? d.getUTCSeconds() : d.getSeconds()

  const date = `${Y}-${pad(M)}-${pad(D)}`
  const time = showSeconds
    ? `${pad(h)}:${pad(m)}:${pad(s)}`
    : `${pad(h)}:${pad(m)}`

  return <span className={className}>{`${date} ${time}`}</span>
}
