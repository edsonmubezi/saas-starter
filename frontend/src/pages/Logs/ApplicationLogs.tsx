import React, { useMemo, useState } from 'react'
import { useQuery, keepPreviousData } from '@tanstack/react-query'
import { FileText, Eye, Clock, Filter, X, Search, AlertCircle } from 'lucide-react'
import Modal from '../../ui/Modal'
import DataTable, { type Column } from '../../ui/DataTable'
import { useAuth } from '../../state/AuthContext'

import {
  getApplicationLogs,
  getLogStats,
  type ApplicationLog,
  type LogsResult,
  type LogStats,
  LOG_LEVELS,
  LOG_CATEGORIES,
} from '../../utils/applogs'

export default function ApplicationLogsPage() {
  const { user } = useAuth()
  const perms: string[] = Array.isArray((user as any)?.permissions) ? (user as any).permissions : []
  const canView = perms.some(p => p.includes('admin') || p.includes('logs') || p.includes('system'))

  // Local state
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [viewing, setViewing] = useState<ApplicationLog | null>(null)
  const [showFilters, setShowFilters] = useState(false)

  // Filters
  const [level, setLevel] = useState('')
  const [category, setCategory] = useState('')
  const [search, setSearch] = useState('')
  const [fromDate, setFromDate] = useState('')
  const [toDate, setToDate] = useState('')

  // Query params
  const params = useMemo(() => ({
    page,
    page_size: pageSize,
    level: level || undefined,
    category: category || undefined,
    search: search || undefined,
    from_date: fromDate ? new Date(fromDate).toISOString() : undefined,
    to_date: toDate ? new Date(toDate).toISOString() : undefined,
  }), [page, pageSize, level, category, search, fromDate, toDate])

  const queryKey = useMemo(() => ['applicationLogs', params] as const, [params])

  // Queries
  const { data, isFetching: loading } = useQuery<LogsResult<ApplicationLog>>({
    queryKey,
    queryFn: () => getApplicationLogs(params),
    placeholderData: keepPreviousData,
    enabled: canView,
  })

  const { data: stats } = useQuery<LogStats>({
    queryKey: ['logStats'],
    queryFn: () => getLogStats(24),
    enabled: canView,
    refetchInterval: 60000,
  })

  const logs = data?.logs ?? []
  const total = data?.total ?? 0

  const clearFilters = () => {
    setLevel('')
    setCategory('')
    setSearch('')
    setFromDate('')
    setToDate('')
    setPage(1)
  }

  const hasFilters = level || category || search || fromDate || toDate

  // Columns
  const columns: Column<ApplicationLog>[] = useMemo(
    () => [
      {
        key: 'created_at',
        title: 'Time',
        sortable: true,
        className: 'w-[180px]',
        render: (r) => (
          <div className="flex items-center gap-2">
            <Clock size={14} className="text-foreground/40" />
            <span className="text-xs">
              {r.created_at ? new Date(r.created_at).toLocaleString() : '—'}
            </span>
          </div>
        ),
      },
      {
        key: 'level',
        title: 'Level',
        className: 'w-[100px]',
        render: (r) => {
          const levelInfo = LOG_LEVELS.find(l => l.value === r.level)
          return (
            <span className={`px-2 py-0.5 rounded text-xs font-medium ${levelInfo?.color || 'text-foreground/60'} ${levelInfo?.bgColor || 'bg-foreground/5'}`}>
              {levelInfo?.label || r.level}
            </span>
          )
        },
      },
      {
        key: 'category',
        title: 'Category',
        className: 'w-[120px]',
        render: (r) => {
          const catInfo = LOG_CATEGORIES.find(c => c.value === r.category)
          return (
            <span className="flex items-center gap-1.5 text-xs">
              <span>{catInfo?.icon || '📋'}</span>
              <span>{catInfo?.label || r.category}</span>
            </span>
          )
        },
      },
      {
        key: 'message',
        title: 'Message',
        render: (r) => (
          <div className="max-w-[400px]">
            <div className="text-sm truncate">{r.message}</div>
            {r.error && (
              <div className="text-xs text-rose-400 truncate mt-0.5 flex items-center gap-1">
                <AlertCircle size={12} />
                {r.error}
              </div>
            )}
          </div>
        ),
      },
      {
        key: 'duration',
        title: 'Duration',
        className: 'w-[80px]',
        render: (r) => (
          <span className="text-xs text-foreground/60">
            {r.duration_ms ? `${r.duration_ms}ms` : '—'}
          </span>
        ),
      },
      {
        key: 'actions',
        title: '',
        className: 'text-right w-[60px]',
        render: (o) => (
          <button
            className="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-foreground/10"
            title="View Details"
            onClick={() => setViewing(o)}
          >
            <Eye size={16} />
          </button>
        ),
      },
    ],
    []
  )

  // No permission UI
  if (!canView) {
    return (
      <div className="card p-6 bg-foreground/5 border-foreground/10">
        <div className="text-sm text-foreground/80 font-medium">
          You do not have permission to view application logs.
        </div>
      </div>
    )
  }

  return (
    <div className="grid gap-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold flex items-center gap-2">
            <FileText className="text-brand-400" />
            Application Logs
          </h1>
          <p className="text-sm text-foreground/60 mt-1">
            View system logs and debug information
          </p>
        </div>
        <button
          onClick={() => setShowFilters(!showFilters)}
          className={`btn btn-secondary flex items-center gap-2 ${hasFilters ? 'ring-2 ring-brand-400' : ''}`}
        >
          <Filter size={16} />
          Filters
          {hasFilters && <span className="text-xs bg-brand-500 px-1.5 rounded">Active</span>}
        </button>
      </div>

      {/* Stats Summary */}
      {stats && (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
          {LOG_LEVELS.map(lvl => {
            const count = stats.logs_by_level?.[lvl.value] || 0
            return (
              <div key={lvl.value} className="card bg-foreground/5 border-foreground/10 py-3">
                <div className="text-center">
                  <div className={`text-2xl font-bold ${lvl.color}`}>{count}</div>
                  <div className="text-xs text-foreground/60">{lvl.label}</div>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Filters */}
      {showFilters && (
        <div className="card bg-foreground/5 border-foreground/10">
          <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4">
            <div>
              <label className="block text-xs text-foreground/60 mb-1">Level</label>
              <select
                value={level}
                onChange={(e) => { setLevel(e.target.value); setPage(1) }}
                className="input w-full"
              >
                <option value="">All Levels</option>
                {LOG_LEVELS.map(l => (
                  <option key={l.value} value={l.value}>{l.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-xs text-foreground/60 mb-1">Category</label>
              <select
                value={category}
                onChange={(e) => { setCategory(e.target.value); setPage(1) }}
                className="input w-full"
              >
                <option value="">All Categories</option>
                {LOG_CATEGORIES.map(c => (
                  <option key={c.value} value={c.value}>{c.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-xs text-foreground/60 mb-1">Search</label>
              <div className="relative">
                <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-foreground/40" />
                <input
                  type="text"
                  value={search}
                  onChange={(e) => { setSearch(e.target.value); setPage(1) }}
                  placeholder="Search messages..."
                  className="input w-full pl-9"
                />
              </div>
            </div>
            <div>
              <label className="block text-xs text-foreground/60 mb-1">From Date</label>
              <input
                type="datetime-local"
                value={fromDate}
                onChange={(e) => { setFromDate(e.target.value); setPage(1) }}
                className="input w-full"
              />
            </div>
            <div>
              <label className="block text-xs text-foreground/60 mb-1">To Date</label>
              <input
                type="datetime-local"
                value={toDate}
                onChange={(e) => { setToDate(e.target.value); setPage(1) }}
                className="input w-full"
              />
            </div>
          </div>
          {hasFilters && (
            <div className="mt-4 pt-4 border-t border-foreground/10">
              <button onClick={clearFilters} className="text-sm text-foreground/60 hover:text-foreground flex items-center gap-1">
                <X size={14} /> Clear all filters
              </button>
            </div>
          )}
        </div>
      )}

      {/* Summary */}
      <div className="card flex flex-wrap items-center gap-4 bg-foreground/5 border-foreground/10">
        <div className="text-sm text-foreground/60">
          {total} log entr{total !== 1 ? 'ies' : 'y'} found
        </div>
      </div>

      {/* Table */}
      <DataTable
        columns={columns}
        rows={logs}
        loading={loading}
        showIndex
        emptyText="No logs found."
        pagination={{
          page,
          pageSize,
          total,
          onPageChange: setPage,
          onPageSizeChange: (s) => {
            setPageSize(s)
            setPage(1)
          },
        }}
      />

      {/* View modal */}
      <Modal
        open={!!viewing}
        onClose={() => setViewing(null)}
        title="Log Details"
        size="lg"
      >
        {viewing && (
          <div className="space-y-4 text-sm">
            {/* Basic Info */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <div>
                <span className="text-foreground/60 block mb-1">Level</span>
                {(() => {
                  const levelInfo = LOG_LEVELS.find(l => l.value === viewing.level)
                  return (
                    <span className={`px-2 py-0.5 rounded text-xs ${levelInfo?.color} ${levelInfo?.bgColor}`}>
                      {levelInfo?.label || viewing.level}
                    </span>
                  )
                })()}
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Category</span>
                {(() => {
                  const catInfo = LOG_CATEGORIES.find(c => c.value === viewing.category)
                  return (
                    <span className="flex items-center gap-1.5">
                      <span>{catInfo?.icon}</span>
                      <span>{catInfo?.label || viewing.category}</span>
                    </span>
                  )
                })()}
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Duration</span>
                <span>{viewing.duration_ms ? `${viewing.duration_ms}ms` : '—'}</span>
              </div>
            </div>

            <div>
              <span className="text-foreground/60 block mb-1">Message</span>
              <div className="p-3 bg-foreground/5 rounded-lg font-mono text-xs whitespace-pre-wrap">
                {viewing.message}
              </div>
            </div>

            {viewing.error && (
              <div>
                <span className="text-rose-400 block mb-1">Error</span>
                <div className="p-3 bg-rose-500/10 rounded-lg font-mono text-xs text-rose-300">
                  {viewing.error}
                </div>
              </div>
            )}

            {viewing.stack_trace && (
              <div>
                <span className="text-foreground/60 block mb-1">Stack Trace</span>
                <pre className="p-3 bg-foreground/5 rounded-lg text-xs overflow-x-auto max-h-[200px]">
                  {viewing.stack_trace}
                </pre>
              </div>
            )}

            {/* Trace IDs */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <span className="text-foreground/60 block mb-1">Trace ID</span>
                <span className="font-mono text-xs">{viewing.trace_id || '—'}</span>
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Request ID</span>
                <span className="font-mono text-xs">{viewing.request_id || '—'}</span>
              </div>
            </div>

            {/* Metadata */}
            {viewing.metadata && Object.keys(viewing.metadata).length > 0 && (
              <div>
                <span className="text-foreground/60 block mb-1">Metadata</span>
                <pre className="p-3 bg-foreground/5 rounded-lg text-xs overflow-x-auto max-h-[150px]">
                  {JSON.stringify(viewing.metadata, null, 2)}
                </pre>
              </div>
            )}

            {/* Footer */}
            <div className="pt-2 border-t border-foreground/10">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
                <div>
                  <span className="text-foreground/40 block">Log ID</span>
                  <span className="font-mono">{viewing.id}</span>
                </div>
                <div>
                  <span className="text-foreground/40 block">Created At</span>
                  <span>{viewing.created_at ? new Date(viewing.created_at).toLocaleString() : '—'}</span>
                </div>
              </div>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
