import React, { useEffect, useMemo, useRef, useState } from 'react'
import { ChevronDown, Search, Check } from 'lucide-react'
import clsx from 'clsx'

export type SSOption = { value: string; label: string; meta?: any }

type Props = {
  value: string
  onChange: (v: string) => void
  options: SSOption[]
  placeholder?: string
  loading?: boolean
  disabled?: boolean
  searchable?: boolean
  className?: string
  panelClassName?: string
  emptyText?: string
}

export default function SearchableSelect({
  value,
  onChange,
  options,
  placeholder = 'Select...',
  loading,
  disabled,
  searchable = true,
  className,
  panelClassName,
  emptyText = 'No matches',
}: Props) {
  const [open, setOpen] = useState(false)
  const [term, setTerm] = useState('')
  const [activeIdx, setActiveIdx] = useState<number>(-1)
  const btnRef = useRef<HTMLButtonElement>(null)
  const panelRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const selected = useMemo(() => options.find(o => o.value === value), [options, value])
  const filtered = useMemo(() => {
    const t = term.trim().toLowerCase()
    if (!t) return options
    return options.filter(o => o.label.toLowerCase().includes(t))
  }, [options, term])

  useEffect(() => {
    function onDocClick(e: MouseEvent) {
      if (!panelRef.current || !btnRef.current) return
      const node = e.target as Node
      if (!panelRef.current.contains(node) && !btnRef.current.contains(node)) setOpen(false)
    }
    document.addEventListener('click', onDocClick)
    return () => document.removeEventListener('click', onDocClick)
  }, [])

  useEffect(() => {
    if (open) {
      setTerm('')
      setActiveIdx(-1)
      setTimeout(() => inputRef.current?.focus(), 0)
    }
  }, [open])

  return (
    <div className={clsx('relative', className)}>
      <button
        ref={btnRef}
        type="button"
        disabled={disabled}
        onClick={() => setOpen(o => !o)}
        className={clsx(
          'w-full flex items-center justify-between gap-2 !text-left rounded-xl px-3 py-2',
          'bg-surface-input text-foreground border',
          open ? 'border-brand-500' : 'border-foreground/20',
          'transition outline-none',
          'focus-visible:ring-2 focus-visible:ring-brand-500/35 focus-visible:border-brand-500',
          disabled && 'opacity-60 cursor-not-allowed'
        )}
        aria-haspopup="listbox"
        aria-expanded={open}
      >
        <span className={clsx(!selected && 'text-foreground/50')}>
          {selected ? selected.label : placeholder}
        </span>
        <ChevronDown
          size={16}
          className={clsx('shrink-0 text-foreground/70 transition-transform', open && 'rotate-180')}
        />
      </button>

      {open && (
        <div
          ref={panelRef}
          className={clsx(
            'absolute z-50 mt-2 w-full rounded-xl overflow-hidden',
            'border border-foreground/12 bg-surface-primary shadow-xl ring-1 ring-black/20 backdrop-blur-sm',
            panelClassName
          )}
          role="listbox"
        >
          {searchable && (
            <div className="p-2 border-b border-foreground/10">
              <div className="relative">
                <input
                  ref={inputRef}
                  value={term}
                  onChange={e => { setTerm(e.target.value); setActiveIdx(-1) }}
                  onKeyDown={e => {
                    if (e.key === 'ArrowDown') {
                      e.preventDefault()
                      setActiveIdx(i => Math.min((i < 0 ? 0 : i + 1), filtered.length - 1))
                    } else if (e.key === 'ArrowUp') {
                      e.preventDefault()
                      setActiveIdx(i => Math.max(i - 1, 0))
                    } else if (e.key === 'Home') {
                      e.preventDefault(); setActiveIdx(0)
                    } else if (e.key === 'End') {
                      e.preventDefault(); setActiveIdx(filtered.length - 1)
                    } else if (e.key === 'Enter') {
                      e.preventDefault()
                      const pick = filtered[activeIdx]
                      if (pick) { onChange(pick.value); setOpen(false) }
                    } else if (e.key === 'Escape' || e.key === 'Tab') {
                      setOpen(false)
                    }
                  }}
                  placeholder="Search..."
                  className="w-full rounded-lg border border-foreground/15 bg-surface-input text-foreground placeholder:text-foreground/40 px-3 py-2 pr-8 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-500/25"
                />
                <Search size={16} className="absolute right-2 top-1/2 -translate-y-1/2 text-foreground/50" />
              </div>
            </div>
          )}

          <div className="max-h-64 overflow-y-auto py-1">
            {loading ? (
              <div className="px-3 py-2 text-sm text-foreground/70">Loading...</div>
            ) : filtered.length === 0 ? (
              <div className="px-3 py-2 text-sm text-foreground/70">{emptyText}</div>
            ) : (
              filtered.map((o, i) => {
                const isSel = o.value === value
                const active = i === activeIdx
                return (
                  <button
                    key={o.value}
                    type="button"
                    onMouseEnter={() => setActiveIdx(i)}
                    onClick={() => { onChange(o.value); setOpen(false) }}
                    className={clsx(
                      'w-full flex items-center justify-between gap-2 text-left px-3 py-2 text-sm transition',
                      active ? 'bg-brand-500/15' : 'hover:bg-foreground/5',
                      isSel ? 'font-semibold text-foreground' : 'text-foreground/90'
                    )}
                    role="option"
                    aria-selected={isSel}
                  >
                    <span className="truncate">{o.label}</span>
                    {isSel && <Check size={16} className="opacity-90" />}
                  </button>
                )
              })
            )}
          </div>
        </div>
      )}
    </div>
  )
}
