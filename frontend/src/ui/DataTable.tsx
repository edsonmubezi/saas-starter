import React from 'react'
import clsx from 'clsx'
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react'

export type Column<T> = {
  key: string
  title: string
  sortable?: boolean
  className?: string
  width?: string | number
  render?: (row: T, idx: number) => React.ReactNode
}

export type Pagination = {
  page: number
  pageSize: number
  total?: number
  onPageChange: (page: number) => void
  onPageSizeChange?: (size: number) => void
}

export type SortState = {
  sortBy?: string
  sortDir?: 'asc' | 'desc'
  onSortChange?: (key: string) => void
}

export default function DataTable<T>({
  columns,
  rows = [],
  loading,
  pagination,
  sort,
  showIndex = true,
  emptyText = 'No records found.',
  onRowClick,
}: {
  columns: Column<T>[]
  rows?: T[] | null
  loading?: boolean
  pagination?: Pagination
  sort?: SortState
  showIndex?: boolean
  emptyText?: string
  onRowClick?: (row: T, idx: number) => void
}) {
  const safeRows: T[] = Array.isArray(rows) ? rows : []
  const page = pagination?.page ?? 1
  const pageSize = pagination?.pageSize ?? (safeRows.length || 10)
  const total = pagination?.total ?? safeRows.length
  const totalPages = Math.ceil(total / pageSize) || 1
  const startIndex = (page - 1) * pageSize
  const startRecord = total > 0 ? startIndex + 1 : 0
  const endRecord = Math.min(startIndex + safeRows.length, total)

  return (
    <div className="rounded-xl border border-foreground/10 bg-surface-elevated overflow-hidden shadow-lg">
      <div className="overflow-x-auto custom-scrollbar">
        <table className="w-full text-sm min-w-full">
          <colgroup>
            {showIndex && <col style={{ width: '56px' }} />}
            {columns.map((col) => (
              <col key={col.key} style={col.width ? { width: col.width as any } : undefined} />
            ))}
          </colgroup>

          <thead>
            <tr className="bg-gradient-to-r from-foreground/[0.03] to-foreground/[0.06] border-b border-foreground/10">
              {showIndex && (
                <th className="px-4 py-4 text-center text-xs font-semibold text-foreground/50 uppercase tracking-wider">
                  #
                </th>
              )}
              {columns.map((col) => {
                const isActive = sort?.sortBy === col.key
                const arrow = isActive ? (sort?.sortDir === 'asc' ? '\u2191' : '\u2193') : ''
                return (
                  <th
                    key={col.key}
                    style={col.width ? { width: col.width } : undefined}
                    className={clsx(
                      'px-4 py-4 text-left text-xs font-semibold text-foreground/70 uppercase tracking-wider',
                      col.className,
                      col.sortable && 'cursor-pointer hover:text-foreground hover:bg-foreground/5 transition-colors'
                    )}
                    onClick={() => col.sortable && sort?.onSortChange?.(col.key)}
                    aria-sort={isActive ? (sort?.sortDir === 'asc' ? 'ascending' : 'descending') : 'none'}
                  >
                    <span className="inline-flex items-center gap-2">
                      {col.title}
                      {col.sortable && (
                        <span className={clsx('text-xs transition-opacity', isActive ? 'opacity-100 text-blue-400' : 'opacity-40')}>
                          {arrow || '\u2195'}
                        </span>
                      )}
                    </span>
                  </th>
                )
              })}
            </tr>
          </thead>

          <tbody className="divide-y divide-foreground/5">
            {loading &&
              Array.from({ length: Math.min(5, pageSize) }).map((_, i) => (
                <tr key={`skeleton-${i}`} className="animate-pulse">
                  {showIndex && (
                    <td className="px-4 py-4">
                      <div className="h-4 w-8 mx-auto bg-foreground/10 rounded" />
                    </td>
                  )}
                  {columns.map((c) => (
                    <td key={c.key} className={clsx('px-4 py-4', c.className)}>
                      <div className="h-4 w-3/4 bg-foreground/10 rounded" />
                    </td>
                  ))}
                </tr>
              ))}

            {!loading &&
              safeRows.map((row, i) => (
                <tr
                  key={startIndex + i}
                  className={clsx('hover:bg-foreground/[0.03] transition-colors group', onRowClick && 'cursor-pointer')}
                  onClick={() => onRowClick?.(row, i)}
                >
                  {showIndex && (
                    <td className="px-4 py-4 text-center text-foreground/40 font-mono text-xs">
                      {startIndex + i + 1}
                    </td>
                  )}
                  {columns.map((col) => (
                    <td key={col.key} className={clsx('px-4 py-4 text-foreground/80', col.className)}>
                      {col.render ? col.render(row, i) : (row as any)[col.key]}
                    </td>
                  ))}
                </tr>
              ))}

            {!loading && safeRows.length === 0 && (
              <tr>
                <td className="px-4 py-12 text-center" colSpan={columns.length + (showIndex ? 1 : 0)}>
                  <div className="flex flex-col items-center gap-2">
                    <div className="w-12 h-12 rounded-full bg-foreground/5 flex items-center justify-center">
                      <svg className="w-6 h-6 text-foreground/30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                      </svg>
                    </div>
                    <p className="text-foreground/50 text-sm">{emptyText}</p>
                  </div>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {pagination && (
        <div className="flex flex-col sm:flex-row items-center justify-between gap-3 px-4 py-3 bg-foreground/[0.02] border-t border-foreground/10">
          <div className="text-xs text-foreground/50">
            {total > 0 ? (
              <>
                Showing <span className="text-foreground/70 font-medium">{startRecord}</span> to{' '}
                <span className="text-foreground/70 font-medium">{endRecord}</span> of{' '}
                <span className="text-foreground/70 font-medium">{total}</span> results
              </>
            ) : (
              'No results'
            )}
          </div>

          <div className="flex items-center gap-3">
            {pagination.onPageSizeChange && (
              <select
                className="bg-foreground/5 border border-foreground/10 rounded-lg px-3 py-1.5 text-xs text-foreground/80 focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-500/50 transition-all cursor-pointer hover:bg-foreground/10"
                value={pagination.pageSize}
                onChange={(e) => pagination.onPageSizeChange?.(Number(e.target.value))}
              >
                {[10, 20, 50, 100].map((n) => (
                  <option key={n} value={n}>
                    {n} rows
                  </option>
                ))}
              </select>
            )}

            <div className="flex items-center gap-1">
              <button className="p-1.5 rounded-lg text-foreground/60 hover:text-foreground hover:bg-foreground/10 disabled:opacity-30 disabled:cursor-not-allowed transition-all" onClick={() => pagination.onPageChange(1)} disabled={page <= 1} title="First page">
                <ChevronsLeft className="w-4 h-4" />
              </button>
              <button className="p-1.5 rounded-lg text-foreground/60 hover:text-foreground hover:bg-foreground/10 disabled:opacity-30 disabled:cursor-not-allowed transition-all" onClick={() => pagination.onPageChange(Math.max(1, page - 1))} disabled={page <= 1} title="Previous page">
                <ChevronLeft className="w-4 h-4" />
              </button>
              <div className="px-3 py-1.5 text-xs text-foreground/70 min-w-[80px] text-center">
                Page <span className="text-foreground font-medium">{page}</span> of{' '}
                <span className="text-foreground font-medium">{totalPages}</span>
              </div>
              <button className="p-1.5 rounded-lg text-foreground/60 hover:text-foreground hover:bg-foreground/10 disabled:opacity-30 disabled:cursor-not-allowed transition-all" onClick={() => pagination.onPageChange(page + 1)} disabled={page >= totalPages} title="Next page">
                <ChevronRight className="w-4 h-4" />
              </button>
              <button className="p-1.5 rounded-lg text-foreground/60 hover:text-foreground hover:bg-foreground/10 disabled:opacity-30 disabled:cursor-not-allowed transition-all" onClick={() => pagination.onPageChange(totalPages)} disabled={page >= totalPages} title="Last page">
                <ChevronsRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
