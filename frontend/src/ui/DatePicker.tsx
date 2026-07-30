import { useState, useRef, useEffect, useCallback } from 'react'
import { ChevronLeft, ChevronRight, Calendar } from 'lucide-react'
import clsx from 'clsx'

const MONTHS = [
  'January','February','March','April','May','June',
  'July','August','September','October','November','December',
]
const DAYS = ['Su','Mo','Tu','We','Th','Fr','Sa']

function pad(n: number) { return n.toString().padStart(2, '0') }
function toYMD(d: Date) { return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}` }
function parseYMD(s: string): Date | null {
  if (!s) return null
  const [y,m,d] = s.split('-').map(Number)
  if (!y || !m || !d) return null
  return new Date(y, m-1, d)
}

function isSameDay(a: Date, b: Date) {
  return a.getFullYear()===b.getFullYear() && a.getMonth()===b.getMonth() && a.getDate()===b.getDate()
}

interface DatePickerProps {
  value: string            // yyyy-mm-dd
  onChange: (val: string) => void
  min?: string             // yyyy-mm-dd — dates before this are disabled
  label?: string
  error?: string
  placeholder?: string
  className?: string
}

export default function DatePicker({ value, onChange, min, label, error, placeholder, className }: DatePickerProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  const selected = parseYMD(value)
  const minDate = parseYMD(min ?? '')

  // Calendar view state
  const [viewYear, setViewYear] = useState(() => selected?.getFullYear() ?? new Date().getFullYear())
  const [viewMonth, setViewMonth] = useState(() => selected?.getMonth() ?? new Date().getMonth())

  // Sync view when value changes externally
  useEffect(() => {
    if (selected) {
      setViewYear(selected.getFullYear())
      setViewMonth(selected.getMonth())
    }
  }, [value]) // eslint-disable-line react-hooks/exhaustive-deps

  // Close on outside click
  useEffect(() => {
    if (!open) return
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  const prevMonth = useCallback(() => {
    setViewMonth(m => { if (m === 0) { setViewYear(y => y-1); return 11 } return m-1 })
  }, [])
  const nextMonth = useCallback(() => {
    setViewMonth(m => { if (m === 11) { setViewYear(y => y+1); return 0 } return m+1 })
  }, [])

  // Build calendar grid
  const firstDay = new Date(viewYear, viewMonth, 1).getDay()
  const daysInMonth = new Date(viewYear, viewMonth+1, 0).getDate()
  const prevMonthDays = new Date(viewYear, viewMonth, 0).getDate()

  const cells: { day: number; month: number; year: number; outside: boolean }[] = []
  // Leading days from previous month
  for (let i = firstDay-1; i >= 0; i--) {
    const d = prevMonthDays - i
    const pm = viewMonth === 0 ? 11 : viewMonth-1
    const py = viewMonth === 0 ? viewYear-1 : viewYear
    cells.push({ day: d, month: pm, year: py, outside: true })
  }
  // Current month
  for (let d = 1; d <= daysInMonth; d++) {
    cells.push({ day: d, month: viewMonth, year: viewYear, outside: false })
  }
  // Trailing days
  const remaining = 42 - cells.length
  for (let d = 1; d <= remaining; d++) {
    const nm = viewMonth === 11 ? 0 : viewMonth+1
    const ny = viewMonth === 11 ? viewYear+1 : viewYear
    cells.push({ day: d, month: nm, year: ny, outside: true })
  }

  const isDisabled = (cell: typeof cells[0]) => {
    if (!minDate) return false
    const d = new Date(cell.year, cell.month, cell.day)
    return d < minDate
  }

  const isSelected = (cell: typeof cells[0]) => {
    if (!selected) return false
    const d = new Date(cell.year, cell.month, cell.day)
    return isSameDay(d, selected)
  }

  const isToday = (cell: typeof cells[0]) => {
    const now = new Date()
    return cell.day === now.getDate() && cell.month === now.getMonth() && cell.year === now.getFullYear()
  }

  const selectDay = (cell: typeof cells[0]) => {
    if (isDisabled(cell)) return
    const d = new Date(cell.year, cell.month, cell.day)
    onChange(toYMD(d))
    setOpen(false)
  }

  const displayValue = selected
    ? `${selected.getDate()}.${MONTHS[selected.getMonth()].slice(0,3)}.${selected.getFullYear()}`
    : ''

  return (
    <div ref={ref} className={clsx('relative', className)}>
      {label && <label className="block text-sm mb-1 text-neutral-300">{label}</label>}

      {/* Trigger */}
      <button
        type="button"
        onClick={() => setOpen(o => !o)}
        className={clsx(
          'w-full flex items-center justify-between gap-2 bg-surface-input text-neutral-100 border rounded-lg px-3 py-2 text-sm text-left focus:outline-none focus:ring-2 transition-colors',
          error ? 'border-red-500 focus:ring-red-500/40' : 'border-foreground/10 focus:ring-white/20',
          !displayValue && 'text-neutral-500',
        )}
      >
        <span>{displayValue || placeholder || 'Select date'}</span>
        <Calendar className="w-4 h-4 text-neutral-400 shrink-0" />
      </button>
      {error && <p className="mt-1 text-xs text-red-300">{error}</p>}

      {/* Dropdown calendar */}
      {open && (
        <div className="absolute z-50 mt-1 w-[280px] rounded-xl border border-foreground/10 bg-surface-elevated shadow-2xl p-3 animate-in fade-in slide-in-from-top-2">
          {/* Header */}
          <div className="flex items-center justify-between mb-2">
            <button type="button" onClick={prevMonth} className="p-1 rounded-md hover:bg-foreground/10 text-neutral-300 transition-colors">
              <ChevronLeft className="w-4 h-4" />
            </button>
            <span className="text-sm font-semibold text-neutral-100">
              {MONTHS[viewMonth]} {viewYear}
            </span>
            <button type="button" onClick={nextMonth} className="p-1 rounded-md hover:bg-foreground/10 text-neutral-300 transition-colors">
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>

          {/* Day headers */}
          <div className="grid grid-cols-7 mb-1">
            {DAYS.map(d => (
              <div key={d} className="text-center text-[11px] font-medium text-neutral-500 py-1">{d}</div>
            ))}
          </div>

          {/* Day grid */}
          <div className="grid grid-cols-7">
            {cells.map((cell, i) => {
              const disabled = isDisabled(cell)
              const sel = isSelected(cell)
              const today = isToday(cell)
              return (
                <button
                  key={i}
                  type="button"
                  disabled={disabled}
                  onClick={() => selectDay(cell)}
                  className={clsx(
                    'h-8 w-full rounded-md text-xs font-medium transition-colors',
                    cell.outside && 'text-neutral-600',
                    !cell.outside && !disabled && !sel && 'text-neutral-200 hover:bg-foreground/10',
                    disabled && 'text-neutral-700 cursor-not-allowed',
                    sel && 'bg-blue-600 text-white hover:bg-blue-500',
                    today && !sel && !disabled && 'ring-1 ring-blue-400/60 text-blue-400',
                  )}
                >
                  {cell.day}
                </button>
              )
            })}
          </div>

          {/* Footer */}
          <div className="flex items-center justify-between mt-2 pt-2 border-t border-foreground/10">
            <button
              type="button"
              onClick={() => { onChange(''); setOpen(false) }}
              className="text-xs text-neutral-400 hover:text-neutral-200 transition-colors"
            >
              Clear
            </button>
            <button
              type="button"
              onClick={() => {
                const now = new Date()
                const tomorrow = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1)
                const target = minDate && tomorrow < minDate ? minDate : tomorrow
                onChange(toYMD(target))
                setOpen(false)
              }}
              className="text-xs text-blue-400 hover:text-blue-300 font-medium transition-colors"
            >
              Tomorrow
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
