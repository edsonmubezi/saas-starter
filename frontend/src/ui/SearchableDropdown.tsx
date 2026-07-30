// src/ui/SearchableDropdown.tsx
import React, { useEffect, useState, useRef } from 'react'

export type SearchableDropdownProps<T> = {
  label: string
  value?: string
  onChange: (v: string) => void

  /** Fetcher returning an array of items */
  loader: () => Promise<T[]>

  /** How to read value/label from each item */
  getValue: (item: T) => string
  getLabel: (item: T) => string

  /** UI bits */
  placeholder?: string
  disabled?: boolean
  id?: string
  name?: string

  /** Validation/error */
  error?: string

  /** Cache key to control when to refetch */
  cacheKey?: string | number
}

export default function SearchableDropdown<T>({
  label,
  value,
  onChange,
  loader,
  getValue,
  getLabel,
  placeholder = 'Select…',
  disabled,
  id,
  name,
  error,
  cacheKey,
}: SearchableDropdownProps<T>) {
  const [opts, setOpts] = useState<T[]>([])
  const [loading, setLoading] = useState(false)
  const [search, setSearch] = useState('')
  const [isOpen, setIsOpen] = useState(false)
  const wrapperRef = useRef<HTMLDivElement>(null)

  // Load options
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

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  // Filter options based on search
  const filteredOpts = opts.filter(o => {
    if (!search) return true
    const label = getLabel(o).toLowerCase()
    return label.includes(search.toLowerCase())
  })

  // Get selected item label
  const selectedItem = opts.find(o => getValue(o) === value)
  const selectedLabel = selectedItem ? getLabel(selectedItem) : placeholder

  const handleSelect = (item: T) => {
    onChange(getValue(item))
    setIsOpen(false)
    setSearch('')
  }

  const ringErr = error ? ' ring-1 ring-rose-400' : ''

  return (
    <div className="flex flex-col gap-1 w-full" ref={wrapperRef}>
      <label htmlFor={id} className="text-sm text-foreground/70">{label}</label>

      <div className="relative">
        {/* Selected value display / trigger button */}
        <button
          type="button"
          id={id}
          className={`w-full text-left appearance-none rounded border border-foreground/15 bg-surface-input px-3 py-2 text-foreground outline-none ring-0 hover:bg-surface-primary focus:border-blue-400/40 focus:ring-2 focus:ring-blue-400/30 disabled:opacity-60 disabled:cursor-not-allowed${ringErr}`}
          onClick={() => !disabled && !loading && setIsOpen(!isOpen)}
          disabled={disabled || loading}
          aria-invalid={!!error}
          aria-describedby={error ? `${id || name}-error` : undefined}
        >
          <div className="flex items-center justify-between">
            <span className={!value ? 'text-foreground/50' : ''}>
              {loading ? 'Loading…' : selectedLabel}
            </span>
            <svg
              className={`w-4 h-4 transition-transform ${isOpen ? 'rotate-180' : ''}`}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </div>
        </button>

        {/* Dropdown panel */}
        {isOpen && !loading && (
          <div className="absolute z-50 mt-1 w-full bg-surface-primary border border-foreground/15 rounded-lg shadow-xl max-h-[300px] flex flex-col">
            {/* Search input */}
            <div className="p-2 border-b border-foreground/10">
              <input
                type="text"
                placeholder="Search..."
                className="w-full bg-surface-input border border-foreground/10 rounded px-3 py-1.5 text-sm text-foreground outline-none focus:border-blue-400/40 focus:ring-1 focus:ring-blue-400/30"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                onClick={(e) => e.stopPropagation()}
              />
            </div>

            {/* Options list */}
            <div className="overflow-y-auto">
              {filteredOpts.length === 0 ? (
                <div className="px-3 py-2 text-sm text-foreground/50">No results found</div>
              ) : (
                filteredOpts.map((o, i) => {
                  const v = getValue(o)
                  const l = getLabel(o)
                  const isSelected = v === value
                  return (
                    <button
                      key={v || i}
                      type="button"
                      className={`w-full text-left px-3 py-2 text-sm transition-colors ${
                        isSelected
                          ? 'bg-blue-500/20 text-blue-300'
                          : 'hover:bg-foreground/5 text-foreground'
                      }`}
                      onClick={() => handleSelect(o)}
                    >
                      {l}
                    </button>
                  )
                })
              )}
            </div>
          </div>
        )}
      </div>

      {error && (
        <span id={`${id || name}-error`} className="text-xs text-rose-400">
          {error}
        </span>
      )}
    </div>
  )
}
