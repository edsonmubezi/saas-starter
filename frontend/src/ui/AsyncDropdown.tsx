import React, { useEffect, useState } from 'react'

type AsyncDropdownProps<T> = {
  label: string
  value?: string
  onChange: (v: string) => void

  /** Fetcher returning an array of items */
  loader: () => Promise<T[]>

  /** How to read value/label from each item */
  getValue: (item: T) => string
  getLabel: (item: T) => string

  /** UI bits */
  placeholder?: string            // text for the first (all) option
  className?: string              // wrapper class
  selectClassName?: string        // select class
  disabled?: boolean
}

export default function AsyncDropdown<T>({
  label,
  value,
  onChange,
  loader,
  getValue,
  getLabel,
  placeholder = 'All',
  className = 'flex items-center gap-2',
  selectClassName = 'select w-full sm:w-[180px]',
  disabled,
}: AsyncDropdownProps<T>) {
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
  }, [loader])

  return (
    <div className={className}>
      <span className="text-sm text-foreground/70">{label}</span>
      <select
        className={selectClassName}
        value={value || ''}
        onChange={e => onChange(e.target.value)}
        disabled={disabled || loading}
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
    </div>
  )
}
