import React, { useMemo, useState } from 'react'
import { useQuery, keepPreviousData } from '@tanstack/react-query'
import { Globe, Eye, Clock, Filter, X } from 'lucide-react'
import Modal from '../../ui/Modal'
import DataTable, { type Column } from '../../ui/DataTable'
import { useAuth } from '../../state/AuthContext'

import {
  getAccessLogs,
  type AccessLog,
  type LogsResult,
  HTTP_METHODS,
} from '../../utils/applogs'

export default function AccessLogsPage() {
  const { user } = useAuth()
  const perms: string[] = Array.isArray((user as any)?.permissions) ? (user as any).permissions : []
  const canView = perms.some(p => p.includes('admin') || p.includes('logs') || p.includes('system'))

  // Local state
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [viewing, setViewing] = useState<AccessLog | null>(null)
  const [showFilters, setShowFilters] = useState(false)

  // Filters
  const [method, setMethod] = useState('')
  const [path, setPath] = useState('')
  const [fromDate, setFromDate] = useState('')
  const [toDate, setToDate] = useState('')

  // Query params
  const params = useMemo(() => ({
    page,
    page_size: pageSize,
    method: method || undefined,
    path: path || undefined,
    from_date: fromDate ? new Date(fromDate).toISOString() : undefined,
    to_date: toDate ? new Date(toDate).toISOString() : undefined,
  }), [page, pageSize, method, path, fromDate, toDate])

  const queryKey = useMemo(() => ['accessLogs', params] as const, [params])

  // Query
  const { data, isFetching: loading } = useQuery<LogsResult<AccessLog>>({
    queryKey,
    queryFn: () => getAccessLogs(params),
    placeholderData: keepPreviousData,
    enabled: canView,
  })

  const logs = data?.logs ?? []
  const total = data?.total ?? 0

  const clearFilters = () => {
    setMethod('')
    setPath('')
    setFromDate('')
    setToDate('')
    setPage(1)
  }

  const hasFilters = method || path || fromDate || toDate

  const getStatusColor = (status: number) => {
    if (status >= 500) return 'text-rose-400 bg-rose-500/10'
    if (status >= 400) return 'text-amber-400 bg-amber-500/10'
    if (status >= 300) return 'text-blue-400 bg-blue-500/10'
    if (status >= 200) return 'text-emerald-400 bg-emerald-500/10'
    return 'text-foreground/60 bg-foreground/5'
  }

  // Columns
  const columns: Column<AccessLog>[] = useMemo(
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
        key: 'method',
        title: 'Method',
        className: 'w-[80px]',
        render: (r) => {
          const methodInfo = HTTP_METHODS.find(m => m.value === r.method)
          return (
            <span className={`font-mono text-xs font-medium ${methodInfo?.color || 'text-foreground/60'}`}>
              {r.method}
            </span>
          )
        },
      },
      {
        key: 'path',
        title: 'Path',
        render: (r) => (
          <span className="font-mono text-xs text-foreground/80 truncate block max-w-[300px]" title={r.path}>
            {r.path}
          </span>
        ),
      },
      {
        key: 'status_code',
        title: 'Status',
        className: 'w-[80px]',
        render: (r) => (
          <span className={`px-2 py-0.5 rounded text-xs font-medium ${getStatusColor(r.status_code)}`}>
            {r.status_code}
          </span>
        ),
      },
      {
        key: 'duration_ms',
        title: 'Duration',
        className: 'w-[90px]',
        render: (r) => {
          const color = r.duration_ms > 1000 ? 'text-rose-400' :
                       r.duration_ms > 500 ? 'text-amber-400' :
                       'text-foreground/60'
          return <span className={`text-xs ${color}`}>{r.duration_ms}ms</span>
        },
      },
      {
        key: 'ip_address',
        title: 'IP',
        className: 'w-[120px]',
        render: (r) => (
          <span className="font-mono text-xs text-foreground/60">{r.ip_address || '—'}</span>
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
          You do not have permission to view access logs.
        </div>
      </div>
    )
  }

  return (
    <div className="grid gap-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold flex items-center gap-2">
            <Globe className="text-brand-400" />
            Access Logs
          </h1>
          <p className="text-sm text-foreground/60 mt-1">
            HTTP request and response logs
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

      {/* Filters */}
      {showFilters && (
        <div className="card bg-foreground/5 border-foreground/10">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div>
              <label className="block text-xs text-foreground/60 mb-1">Method</label>
              <select
                value={method}
                onChange={(e) => { setMethod(e.target.value); setPage(1) }}
                className="input w-full"
              >
                <option value="">All Methods</option>
                {HTTP_METHODS.map(m => (
                  <option key={m.value} value={m.value}>{m.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-xs text-foreground/60 mb-1">Path (contains)</label>
              <input
                type="text"
                value={path}
                onChange={(e) => { setPath(e.target.value); setPage(1) }}
                placeholder="/api/..."
                className="input w-full font-mono"
              />
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
          {total} request{total !== 1 ? 's' : ''} found
        </div>
      </div>

      {/* Table */}
      <DataTable
        columns={columns}
        rows={logs}
        loading={loading}
        showIndex
        emptyText="No access logs found."
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
        title="Request Details"
        size="lg"
      >
        {viewing && (
          <div className="space-y-4 text-sm">
            {/* Request Info */}
            <div className="p-4 bg-foreground/5 rounded-lg">
              <div className="flex items-center gap-3">
                {(() => {
                  const methodInfo = HTTP_METHODS.find(m => m.value === viewing.method)
                  return (
                    <span className={`font-mono text-sm font-bold ${methodInfo?.color}`}>
                      {viewing.method}
                    </span>
                  )
                })()}
                <span className="font-mono text-sm text-foreground/80 flex-1 truncate">{viewing.path}</span>
                <span className={`px-2 py-0.5 rounded text-xs font-medium ${getStatusColor(viewing.status_code)}`}>
                  {viewing.status_code}
                </span>
              </div>
            </div>

            {/* Performance */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <div>
                <span className="text-foreground/60 block mb-1">Duration</span>
                <span className={`text-lg font-bold ${
                  viewing.duration_ms > 1000 ? 'text-rose-400' :
                  viewing.duration_ms > 500 ? 'text-amber-400' :
                  'text-emerald-400'
                }`}>{viewing.duration_ms}ms</span>
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Request Size</span>
                <span>{viewing.request_size ? `${viewing.request_size} bytes` : '—'}</span>
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Response Size</span>
                <span>{viewing.response_size ? `${viewing.response_size} bytes` : '—'}</span>
              </div>
            </div>

            {/* Client Info */}
            <div className="p-3 bg-foreground/5 rounded-lg">
              <span className="text-foreground/60 block mb-2">Client</span>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <span className="text-foreground/40 text-xs block">IP Address</span>
                  <span className="font-mono">{viewing.ip_address || '—'}</span>
                </div>
                <div>
                  <span className="text-foreground/40 text-xs block">User ID</span>
                  <span>{viewing.user_id || '—'}</span>
                </div>
              </div>
              {viewing.user_agent && (
                <div className="mt-3">
                  <span className="text-foreground/40 text-xs block">User Agent</span>
                  <span className="text-xs text-foreground/70 break-all">{viewing.user_agent}</span>
                </div>
              )}
            </div>

            {/* Error */}
            {viewing.error && (
              <div>
                <span className="text-rose-400 block mb-1">Error</span>
                <div className="p-3 bg-rose-500/10 rounded-lg font-mono text-xs text-rose-300">
                  {viewing.error}
                </div>
              </div>
            )}

            {/* IDs */}
            <div className="pt-2 border-t border-foreground/10">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
                <div>
                  <span className="text-foreground/40 block">Log ID</span>
                  <span className="font-mono">{viewing.id}</span>
                </div>
                <div>
                  <span className="text-foreground/40 block">Request ID</span>
                  <span className="font-mono">{viewing.request_id || '—'}</span>
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
