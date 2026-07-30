import React, { useMemo, useState } from 'react'
import { useQuery, keepPreviousData } from '@tanstack/react-query'
import { ShieldAlert, Eye, Clock, Filter, X } from 'lucide-react'
import Modal from '../../ui/Modal'
import DataTable, { type Column } from '../../ui/DataTable'
import { useAuth } from '../../state/AuthContext'

import {
  getSecurityEvents,
  type SecurityEvent,
  type SecurityEventsResult,
  SECURITY_SEVERITIES,
  SECURITY_CATEGORIES,
  SECURITY_EVENT_TYPES,
} from '../../utils/security'

export default function SecurityEventsPage() {
  const { user } = useAuth()
  const perms: string[] = Array.isArray((user as any)?.permissions) ? (user as any).permissions : []
  const canView = perms.some(p => p.includes('admin') || p.includes('security'))

  // Local state
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [viewing, setViewing] = useState<SecurityEvent | null>(null)
  const [showFilters, setShowFilters] = useState(false)

  // Filters
  const [severity, setSeverity] = useState('')
  const [category, setCategory] = useState('')
  const [eventType, setEventType] = useState('')
  const [fromDate, setFromDate] = useState('')
  const [toDate, setToDate] = useState('')

  // Query params
  const params = useMemo(() => ({
    page,
    page_size: pageSize,
    severity: severity || undefined,
    category: category || undefined,
    event_type: eventType || undefined,
    from: fromDate ? new Date(fromDate).toISOString() : undefined,
    to: toDate ? new Date(toDate).toISOString() : undefined,
  }), [page, pageSize, severity, category, eventType, fromDate, toDate])

  const queryKey = useMemo(() => ['securityEvents', params] as const, [params])

  // Query
  const { data, isFetching: loading } = useQuery<SecurityEventsResult>({
    queryKey,
    queryFn: () => getSecurityEvents(params),
    placeholderData: keepPreviousData,
    enabled: canView,
  })

  const events = data?.events ?? []
  const total = data?.total ?? 0

  const clearFilters = () => {
    setSeverity('')
    setCategory('')
    setEventType('')
    setFromDate('')
    setToDate('')
    setPage(1)
  }

  const hasFilters = severity || category || eventType || fromDate || toDate

  // Columns
  const columns: Column<SecurityEvent>[] = useMemo(
    () => [
      {
        key: 'created_at',
        title: 'Time',
        sortable: true,
        render: (r) => (
          <div className="flex items-center gap-2">
            <Clock size={14} className="text-foreground/40" />
            <span className="text-sm">
              {r.created_at ? new Date(r.created_at).toLocaleString() : '—'}
            </span>
          </div>
        ),
      },
      {
        key: 'event_type',
        title: 'Event',
        render: (r) => {
          const eventInfo = SECURITY_EVENT_TYPES.find(e => e.value === r.event_type)
          return (
            <div>
              <div className="font-medium text-sm">{eventInfo?.label || r.event_type}</div>
              <div className="text-xs text-foreground/50 truncate max-w-[200px]">{r.description}</div>
            </div>
          )
        },
      },
      {
        key: 'category',
        title: 'Category',
        render: (r) => {
          const catInfo = SECURITY_CATEGORIES.find(c => c.value === r.category)
          return (
            <span className="flex items-center gap-1.5 text-sm">
              <span>{catInfo?.icon || '📋'}</span>
              <span>{catInfo?.label || r.category}</span>
            </span>
          )
        },
      },
      {
        key: 'severity',
        title: 'Severity',
        render: (r) => {
          const sevInfo = SECURITY_SEVERITIES.find(s => s.value === r.severity)
          return (
            <span className={`px-2 py-0.5 rounded text-xs ${sevInfo?.color || 'text-foreground/60'} ${sevInfo?.bgColor || 'bg-foreground/5'}`}>
              {sevInfo?.label || r.severity}
            </span>
          )
        },
      },
      {
        key: 'actor',
        title: 'Actor',
        render: (r) => (
          <div>
            <div className="text-sm">{r.actor_name || r.actor_email || '—'}</div>
            {r.ip_address && (
              <div className="text-xs text-foreground/50 font-mono">{r.ip_address}</div>
            )}
          </div>
        ),
      },
      {
        key: 'risk_score',
        title: 'Risk',
        render: (r) => {
          if (!r.risk_score) return <span className="text-foreground/40">—</span>
          const color = r.risk_score >= 80 ? 'text-rose-400' :
                       r.risk_score >= 50 ? 'text-amber-400' :
                       'text-emerald-400'
          return <span className={`font-medium ${color}`}>{r.risk_score}</span>
        },
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
          You do not have permission to view security events.
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
            <ShieldAlert className="text-brand-400" />
            Security Events
          </h1>
          <p className="text-sm text-foreground/60 mt-1">
            Monitor authentication, authorization, and security incidents
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
          <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4">
            <div>
              <label className="block text-xs text-foreground/60 mb-1">Severity</label>
              <select
                value={severity}
                onChange={(e) => { setSeverity(e.target.value); setPage(1) }}
                className="input w-full"
              >
                <option value="">All Severities</option>
                {SECURITY_SEVERITIES.map(s => (
                  <option key={s.value} value={s.value}>{s.label}</option>
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
                {SECURITY_CATEGORIES.map(c => (
                  <option key={c.value} value={c.value}>{c.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-xs text-foreground/60 mb-1">Event Type</label>
              <select
                value={eventType}
                onChange={(e) => { setEventType(e.target.value); setPage(1) }}
                className="input w-full"
              >
                <option value="">All Types</option>
                {SECURITY_EVENT_TYPES.map(et => (
                  <option key={et.value} value={et.value}>{et.label}</option>
                ))}
              </select>
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
          {total} security event{total !== 1 ? 's' : ''} found
        </div>
      </div>

      {/* Table */}
      <DataTable
        columns={columns}
        rows={events}
        loading={loading}
        showIndex
        emptyText="No security events found."
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
        title="Security Event Details"
        size="lg"
      >
        {viewing && (
          <div className="space-y-4 text-sm">
            {/* Event Info */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <span className="text-foreground/60 block mb-1">Event Type</span>
                <span className="font-medium">
                  {SECURITY_EVENT_TYPES.find(e => e.value === viewing.event_type)?.label || viewing.event_type}
                </span>
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Severity</span>
                {(() => {
                  const sevInfo = SECURITY_SEVERITIES.find(s => s.value === viewing.severity)
                  return (
                    <span className={`px-2 py-0.5 rounded text-xs ${sevInfo?.color} ${sevInfo?.bgColor}`}>
                      {sevInfo?.label || viewing.severity}
                    </span>
                  )
                })()}
              </div>
            </div>

            <div>
              <span className="text-foreground/60 block mb-1">Description</span>
              <div className="p-3 bg-foreground/5 rounded-lg">{viewing.description}</div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <span className="text-foreground/60 block mb-1">Category</span>
                {(() => {
                  const catInfo = SECURITY_CATEGORIES.find(c => c.value === viewing.category)
                  return (
                    <span className="flex items-center gap-1.5">
                      <span>{catInfo?.icon}</span>
                      <span>{catInfo?.label || viewing.category}</span>
                    </span>
                  )
                })()}
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Risk Score</span>
                {viewing.risk_score ? (
                  <span className={`font-medium ${
                    viewing.risk_score >= 80 ? 'text-rose-400' :
                    viewing.risk_score >= 50 ? 'text-amber-400' :
                    'text-emerald-400'
                  }`}>{viewing.risk_score}/100</span>
                ) : (
                  <span className="text-foreground/40">—</span>
                )}
              </div>
            </div>

            {/* Actor Info */}
            <div className="p-3 bg-foreground/5 rounded-lg">
              <span className="text-foreground/60 block mb-2">Actor</span>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <span className="text-foreground/40 text-xs block">Name</span>
                  <span>{viewing.actor_name || '—'}</span>
                </div>
                <div>
                  <span className="text-foreground/40 text-xs block">Email</span>
                  <span>{viewing.actor_email || '—'}</span>
                </div>
                <div>
                  <span className="text-foreground/40 text-xs block">IP Address</span>
                  <span className="font-mono">{viewing.ip_address || '—'}</span>
                </div>
                <div>
                  <span className="text-foreground/40 text-xs block">User Agent</span>
                  <span className="text-xs truncate block" title={viewing.user_agent}>
                    {viewing.user_agent || '—'}
                  </span>
                </div>
              </div>
            </div>

            {/* Threat Indicators */}
            {viewing.threat_indicators && viewing.threat_indicators.length > 0 && (
              <div>
                <span className="text-foreground/60 block mb-2">Threat Indicators</span>
                <div className="flex flex-wrap gap-2">
                  {viewing.threat_indicators.map((ind, i) => (
                    <span key={i} className="px-2 py-1 bg-rose-500/10 text-rose-400 rounded text-xs">
                      {ind}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {/* Metadata */}
            {viewing.metadata && Object.keys(viewing.metadata).length > 0 && (
              <div>
                <span className="text-foreground/60 block mb-1">Metadata</span>
                <pre className="p-3 bg-foreground/5 rounded-lg text-xs overflow-x-auto max-h-[150px]">
                  {JSON.stringify(viewing.metadata, null, 2)}
                </pre>
              </div>
            )}

            {/* Timestamps & IDs */}
            <div className="pt-2 border-t border-foreground/10">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
                <div>
                  <span className="text-foreground/40 block">Event ID</span>
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
