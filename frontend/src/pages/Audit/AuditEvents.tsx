import React, { useMemo, useState } from 'react'
import { useQuery, keepPreviousData } from '@tanstack/react-query'
import { FileSearch, Eye, Clock, User, Filter, X } from 'lucide-react'
import Modal from '../../ui/Modal'
import DataTable, { type Column } from '../../ui/DataTable'
import { useAuth } from '../../state/AuthContext'

import {
  getAuditEvents,
  getAuditActions,
  getAuditSeverities,
  getAuditResourceTypes,
  type AuditEvent,
  type AuditEventsResult,
  type AuditMetadataItem,
} from '../../utils/audit'

export default function AuditEventsPage() {
  const { user } = useAuth()
  const perms: string[] = Array.isArray((user as any)?.permissions) ? (user as any).permissions : []
  const canView = perms.some(p => p.includes('admin') || p.includes('audit'))

  // Fetch metadata from backend
  const { data: actions = [] } = useQuery<AuditMetadataItem[]>({
    queryKey: ['audit-actions'],
    queryFn: getAuditActions,
    enabled: canView,
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
  })

  const { data: severities = [] } = useQuery<AuditMetadataItem[]>({
    queryKey: ['audit-severities'],
    queryFn: getAuditSeverities,
    enabled: canView,
    staleTime: 5 * 60 * 1000,
  })

  const { data: resourceTypes = [] } = useQuery<AuditMetadataItem[]>({
    queryKey: ['audit-resource-types'],
    queryFn: getAuditResourceTypes,
    enabled: canView,
    staleTime: 5 * 60 * 1000,
  })

  // Local state
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [viewing, setViewing] = useState<AuditEvent | null>(null)
  const [showFilters, setShowFilters] = useState(false)

  // Filters
  const [resourceType, setResourceType] = useState('')
  const [action, setAction] = useState('')
  const [severity, setSeverity] = useState('')
  const [fromDate, setFromDate] = useState('')
  const [toDate, setToDate] = useState('')

  // Query params
  const params = useMemo(() => ({
    page,
    page_size: pageSize,
    resource_type: resourceType || undefined,
    action: action || undefined,
    severity: severity || undefined,
    from: fromDate ? new Date(fromDate).toISOString() : undefined,
    to: toDate ? new Date(toDate).toISOString() : undefined,
  }), [page, pageSize, resourceType, action, severity, fromDate, toDate])

  const queryKey = useMemo(() => ['auditEvents', params] as const, [params])

  // Query
  const { data, isFetching: loading } = useQuery<AuditEventsResult>({
    queryKey,
    queryFn: () => getAuditEvents(params),
    placeholderData: keepPreviousData,
    enabled: canView,
  })

  const events = data?.events ?? []
  const total = data?.total ?? 0

  const clearFilters = () => {
    setResourceType('')
    setAction('')
    setSeverity('')
    setFromDate('')
    setToDate('')
    setPage(1)
  }

  const hasFilters = resourceType || action || severity || fromDate || toDate

  // Columns
  const columns: Column<AuditEvent>[] = useMemo(
    () => [
      {
        key: 'timestamp',
        title: 'Time',
        sortable: true,
        render: (r) => (
          <div className="flex items-center gap-2">
            <Clock size={14} className="text-foreground/40" />
            <span className="text-sm">
              {r.timestamp ? new Date(r.timestamp).toLocaleString() : '—'}
            </span>
          </div>
        ),
      },
      {
        key: 'actor',
        title: 'Actor',
        render: (r) => (
          <div className="flex items-center gap-2">
            <User size={14} className="text-foreground/40" />
            <div>
              <div className="font-medium text-sm">{r.actor_name || r.actor_email || '—'}</div>
              <div className="text-xs text-foreground/50">{r.actor_type}</div>
            </div>
          </div>
        ),
      },
      {
        key: 'action',
        title: 'Action',
        render: (r) => {
          const actionInfo = actions.find(a => a.value === r.action)
          return (
            <span className={`px-2 py-0.5 rounded text-xs font-medium ${actionInfo?.color || 'text-foreground/60'} bg-foreground/5`}>
              {actionInfo?.label || r.action}
            </span>
          )
        },
      },
      {
        key: 'resource',
        title: 'Resource',
        render: (r) => (
          <div>
            <div className="font-medium text-sm">{r.resource_type}</div>
            <div className="text-xs text-foreground/50">
              {r.resource_name || (r.resource_id ? `ID: ${r.resource_id}` : '—')}
            </div>
          </div>
        ),
      },
      {
        key: 'severity',
        title: 'Severity',
        render: (r) => {
          const sevInfo = severities.find(s => s.value === r.severity)
          return (
            <span className={`px-2 py-0.5 rounded text-xs ${sevInfo?.color || 'text-foreground/60'} bg-foreground/5`}>
              {sevInfo?.label || r.severity}
            </span>
          )
        },
      },
      {
        key: 'ip_address',
        title: 'IP Address',
        render: (r) => (
          <span className="text-xs font-mono text-foreground/60">{r.ip_address || '—'}</span>
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
    [actions, severities]
  )

  // No permission UI
  if (!canView) {
    return (
      <div className="card p-6 bg-foreground/5 border-foreground/10">
        <div className="text-sm text-foreground/80 font-medium">
          You do not have permission to view audit logs.
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
            <FileSearch className="text-brand-400" />
            Audit Logs
          </h1>
          <p className="text-sm text-foreground/60 mt-1">
            Track all changes and activities in the system
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
              <label className="block text-xs text-foreground/60 mb-1">Resource Type</label>
              <select
                value={resourceType}
                onChange={(e) => { setResourceType(e.target.value); setPage(1) }}
                className="input w-full"
              >
                <option value="">All Resources</option>
                {resourceTypes.map(rt => (
                  <option key={rt.value} value={rt.value}>{rt.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-xs text-foreground/60 mb-1">Action</label>
              <select
                value={action}
                onChange={(e) => { setAction(e.target.value); setPage(1) }}
                className="input w-full"
              >
                <option value="">All Actions</option>
                {actions.map(a => (
                  <option key={a.value} value={a.value}>{a.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-xs text-foreground/60 mb-1">Severity</label>
              <select
                value={severity}
                onChange={(e) => { setSeverity(e.target.value); setPage(1) }}
                className="input w-full"
              >
                <option value="">All Severities</option>
                {severities.map(s => (
                  <option key={s.value} value={s.value}>{s.label}</option>
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
          {total} audit event{total !== 1 ? 's' : ''} found
        </div>
      </div>

      {/* Table */}
      <DataTable
        columns={columns}
        rows={events}
        loading={loading}
        showIndex
        emptyText="No audit events found."
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
        title="Audit Event Details"
        size="lg"
      >
        {viewing && (
          <div className="space-y-4 text-sm">
            {/* Basic Info */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <span className="text-foreground/60 block mb-1">Timestamp</span>
                <span>{viewing.timestamp ? new Date(viewing.timestamp).toLocaleString() : '—'}</span>
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Audit ID</span>
                <span className="font-mono text-xs">{viewing.audit_id}</span>
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
                  <span className="text-foreground/40 text-xs block">Type</span>
                  <span>{viewing.actor_type}</span>
                </div>
                <div>
                  <span className="text-foreground/40 text-xs block">ID</span>
                  <span>{viewing.actor_id || '—'}</span>
                </div>
              </div>
            </div>

            {/* Action & Resource */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <span className="text-foreground/60 block mb-1">Action</span>
                {(() => {
                  const actionInfo = actions.find(a => a.value === viewing.action)
                  return (
                    <span className={`px-2 py-0.5 rounded text-xs ${actionInfo?.color || 'text-foreground/60'} bg-foreground/5`}>
                      {actionInfo?.label || viewing.action}
                    </span>
                  )
                })()}
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Severity</span>
                {(() => {
                  const sevInfo = severities.find(s => s.value === viewing.severity)
                  return (
                    <span className={`px-2 py-0.5 rounded text-xs ${sevInfo?.color || 'text-foreground/60'} bg-foreground/5`}>
                      {sevInfo?.label || viewing.severity}
                    </span>
                  )
                })()}
              </div>
            </div>

            {/* Resource */}
            <div className="p-3 bg-foreground/5 rounded-lg">
              <span className="text-foreground/60 block mb-2">Resource</span>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                <div>
                  <span className="text-foreground/40 text-xs block">Type</span>
                  <span>{viewing.resource_type}</span>
                </div>
                <div>
                  <span className="text-foreground/40 text-xs block">ID</span>
                  <span>{viewing.resource_id || '—'}</span>
                </div>
                <div>
                  <span className="text-foreground/40 text-xs block">Name</span>
                  <span>{viewing.resource_name || '—'}</span>
                </div>
              </div>
            </div>

            {/* Request Info */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <span className="text-foreground/60 block mb-1">IP Address</span>
                <span className="font-mono text-xs">{viewing.ip_address || '—'}</span>
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">User Agent</span>
                <span className="text-xs truncate block" title={viewing.user_agent}>
                  {viewing.user_agent || '—'}
                </span>
              </div>
            </div>

            {/* Changes */}
            {viewing.changes && viewing.changes.length > 0 && (
              <div>
                <span className="text-foreground/60 block mb-2">Changes</span>
                <div className="space-y-2 max-h-[200px] overflow-y-auto">
                  {viewing.changes.map((change, i) => (
                    <div key={i} className="p-2 bg-foreground/5 rounded text-xs">
                      <div className="font-medium text-foreground/80">{change.field}</div>
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mt-1">
                        <div>
                          <span className="text-rose-400/60">Old:</span>
                          <pre className="text-rose-300 mt-0.5 overflow-x-auto">
                            {JSON.stringify(change.old_value, null, 2)}
                          </pre>
                        </div>
                        <div>
                          <span className="text-emerald-400/60">New:</span>
                          <pre className="text-emerald-300 mt-0.5 overflow-x-auto">
                            {JSON.stringify(change.new_value, null, 2)}
                          </pre>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Before/After State */}
            {(viewing.before_state || viewing.after_state) && (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {viewing.before_state && (
                  <div>
                    <span className="text-foreground/60 block mb-1">Before State</span>
                    <pre className="p-2 bg-rose-500/10 rounded text-xs overflow-x-auto max-h-[150px]">
                      {JSON.stringify(viewing.before_state, null, 2)}
                    </pre>
                  </div>
                )}
                {viewing.after_state && (
                  <div>
                    <span className="text-foreground/60 block mb-1">After State</span>
                    <pre className="p-2 bg-emerald-500/10 rounded text-xs overflow-x-auto max-h-[150px]">
                      {JSON.stringify(viewing.after_state, null, 2)}
                    </pre>
                  </div>
                )}
              </div>
            )}

            {/* ID */}
            <div className="pt-2 border-t border-foreground/10">
              <span className="text-foreground/40 text-xs block">Event ID</span>
              <span className="font-mono text-xs">{viewing.id}</span>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
