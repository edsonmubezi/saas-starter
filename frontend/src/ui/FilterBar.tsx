import React from 'react'
import { Search, X, SlidersHorizontal } from 'lucide-react'
import clsx from 'clsx'

interface FilterOption {
  value: string
  label: string
}

interface FilterConfig {
  key: string
  label: string
  options: FilterOption[]
  value: string
  onChange: (value: string) => void
  placeholder?: string
}

interface FilterBarProps {
  /** Search input configuration */
  search?: {
    value: string
    onChange: (value: string) => void
    placeholder?: string
  }
  /** Dropdown filters */
  filters?: FilterConfig[]
  /** Clear all filters callback */
  onClear?: () => void
  /** Show clear button */
  showClear?: boolean
  /** Additional content (e.g., custom buttons) */
  children?: React.ReactNode
  /** Custom class name */
  className?: string
}

export default function FilterBar({
  search,
  filters,
  onClear,
  showClear = true,
  children,
  className,
}: FilterBarProps) {
  const hasActiveFilters = filters?.some(f => f.value) || (search?.value && search.value.length > 0)

  return (
    <div className={clsx('bg-foreground/5 border border-foreground/10 rounded-xl p-4 backdrop-blur', className)}>
      <div className="flex flex-wrap items-center gap-3">
        {/* Search Input */}
        {search && (
          <div className="flex-1 min-w-[200px] max-w-md">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground/40" />
              <input
                type="text"
                value={search.value}
                onChange={(e) => search.onChange(e.target.value)}
                placeholder={search.placeholder || 'Search...'}
                className="w-full bg-foreground/5 border border-foreground/10 rounded-lg pl-10 pr-4 py-2.5 text-foreground placeholder:text-foreground/40 outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500/50 transition-all"
              />
              {search.value && (
                <button
                  onClick={() => search.onChange('')}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-foreground/40 hover:text-foreground/60"
                >
                  <X size={14} />
                </button>
              )}
            </div>
          </div>
        )}

        {/* Dropdown Filters */}
        {filters?.map((filter) => (
          <div key={filter.key} className="min-w-[140px]">
            <select
              value={filter.value}
              onChange={(e) => filter.onChange(e.target.value)}
              className="w-full bg-foreground/5 border border-foreground/10 rounded-lg px-3 py-2.5 text-foreground text-sm outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500/50 transition-all appearance-none cursor-pointer"
              style={{
                backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='rgba(255,255,255,0.4)' stroke-width='2'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E")`,
                backgroundRepeat: 'no-repeat',
                backgroundPosition: 'right 0.75rem center',
              }}
            >
              <option value="" className="bg-surface-input">
                {filter.placeholder || `All ${filter.label}`}
              </option>
              {filter.options.map((opt) => (
                <option key={opt.value} value={opt.value} className="bg-surface-input">
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
        ))}

        {/* Custom Content */}
        {children}

        {/* Clear Filters Button */}
        {showClear && hasActiveFilters && onClear && (
          <button
            onClick={onClear}
            className="flex items-center gap-2 px-3 py-2.5 text-sm text-foreground/60 hover:text-foreground hover:bg-foreground/5 rounded-lg transition-all"
          >
            <X size={14} />
            Clear
          </button>
        )}

        {/* Filter Toggle (for mobile) */}
        {filters && filters.length > 0 && (
          <button className="lg:hidden p-2.5 bg-foreground/5 border border-foreground/10 rounded-lg text-foreground/60 hover:text-foreground hover:bg-foreground/10 transition-all">
            <SlidersHorizontal size={18} />
          </button>
        )}
      </div>
    </div>
  )
}

// Export a simple search-only variant
export function SearchBar({
  value,
  onChange,
  placeholder = 'Search...',
  className,
}: {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  className?: string
}) {
  return (
    <div className={clsx('relative', className)}>
      <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground/40" />
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full bg-foreground/5 border border-foreground/10 rounded-lg pl-10 pr-4 py-2.5 text-foreground placeholder:text-foreground/40 outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500/50 transition-all"
      />
      {value && (
        <button
          onClick={() => onChange('')}
          className="absolute right-3 top-1/2 -translate-y-1/2 text-foreground/40 hover:text-foreground/60"
        >
          <X size={14} />
        </button>
      )}
    </div>
  )
}
