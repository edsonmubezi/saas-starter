// src/ui/AsyncDropdownField.tsx
import React, { useEffect, useState } from 'react'

export type AsyncDropdownFieldProps<T> = {
  label: string
  value?: string
  onChange: (v: string) => void

  /** Fetcher returning an array of items */
  loader: () => Promise<T[]>

  /** How to read value/label from each item */
  getValue: (item: T) => string
  getLabel: (item: T) => string

  /** UI bits */
  placeholder?: string        // first option text
  className?: string          // wrapper class
  selectClassName?: string    // select class
  disabled?: boolean
  id?: string
  name?: string

  /** New: validation/error */
  error?: string

  /** Cache key to control when to refetch (prevents unnecessary API calls) */
  cacheKey?: string | number
}

export default function AsyncDropdownField<T>({
  label,
  value,
  onChange,
  loader,
  getValue,
  getLabel,
  placeholder = 'Select…',
  className = 'flex flex-col gap-1 w-full',
  selectClassName = 'bg-dark/5 border border-dark/10 rounded px-3 py-2 outline-none focus:ring-1 focus:ring-blue-500',
  disabled,
  id,
  name,
  error,
  cacheKey,
}: AsyncDropdownFieldProps<T>) {
  const [opts, setOpts] = useState<T[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      setLoading(true)
      try {
        const items = await loader()
        if (!cancelled) setOpts(Array.isArray(items) ? items : [])
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => { cancelled = true }
  }, [cacheKey, loader])

  const ringErr = error ? ' ring-1 ring-rose-400' : ''

  return (
    <div className={className}>
      <label htmlFor={id} className="text-sm text-foreground/70">{label}</label>
      <select
        id={id}
        name={name}
        className={`${selectClassName}${ringErr}`}
        value={value ?? ''}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled || loading}
        aria-invalid={!!error}
        aria-describedby={error ? `${id || name}-error` : undefined}
      >
        <option value="">{loading ? 'Loading…' : placeholder}</option>
        {opts.map((o, i) => {
          const v = getValue(o)
          const l = getLabel(o)
          return (
            <option key={v || i} value={v}>
              {l}
            </option>
          )
        })}
      </select>
      {error && (
        <span id={`${id || name}-error`} className="text-xs text-rose-400">
          {error}
        </span>
      )}
    </div>
  )
}
