import React, { useMemo, useState } from 'react'
import { useQuery, keepPreviousData } from '@tanstack/react-query'
import { History, CheckCircle, XCircle, Clock, Eye } from 'lucide-react'
import Modal from '../../ui/Modal'
import DataTable, { type Column } from '../../ui/DataTable'
import { useAuth } from '../../state/AuthContext'

import {
  getAlertHistory,
  type AlertHistoryResult,
  type AlertHistoryItem,
  SEVERITY_LEVELS,
} from '../../utils/alerting'

export default function AlertHistoryPage() {
  const { user } = useAuth()
  const perms: string[] = Array.isArray((user as any)?.permissions) ? (user as any).permissions : []
  const canView = perms.some(p => p.includes('admin') || p.includes('alerting') || p.includes('security'))

  // Local state
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [viewing, setViewing] = useState<AlertHistoryItem | null>(null)

  // Query params
  const params = useMemo(() => ({ page, page_size: pageSize }), [page, pageSize])
  const queryKey = useMemo(() => ['alertHistory', params] as const, [params])

  // Query
  const { data, isFetching: loading } = useQuery<AlertHistoryResult>({
    queryKey,
    queryFn: () => getAlertHistory(params),
    placeholderData: keepPreviousData,
    enabled: canView,
  })

  const items = data?.items ?? []
  const total = data?.total ?? 0

  // Calculate delivery success
  const getDeliveryStatus = (item: AlertHistoryItem) => {
    const results = item.delivery_results || []
    const successful = results.filter(r => r.success).length
    const total = results.length
    return { successful, total, allSuccess: successful === total && total > 0 }
  }

  // Columns
  const columns: Column<AlertHistoryItem>[] = useMemo(
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
        key: 'alert_title',
        title: 'Alert',
        render: (r) => (
          <div>
            <div className="font-medium">{r.alert_title || r.event_type || '—'}</div>
            <div className="text-xs text-foreground/50 truncate max-w-[200px]">
              {r.alert_message}
            </div>
          </div>
        ),
      },
      {
        key: 'severity',
        title: 'Severity',
        render: (r) => {
          const sevInfo = SEVERITY_LEVELS.find(s => s.value === r.severity)
          return (
            <span className={`px-2 py-0.5 rounded text-xs ${sevInfo?.color || 'text-foreground/60'} bg-foreground/5`}>
              {sevInfo?.label || r.severity}
            </span>
          )
        },
      },
      {
        key: 'channels_notified',
        title: 'Channels',
        render: (r) => (
          <div className="flex flex-wrap gap-1">
            {r.channels_notified?.map((ch) => (
              <span key={ch} className="px-1.5 py-0.5 rounded text-xs bg-foreground/10 capitalize">
                {ch}
              </span>
            ))}
          </div>
        ),
      },
      {
        key: 'delivery',
        title: 'Delivery',
        render: (r) => {
          const { successful, total, allSuccess } = getDeliveryStatus(r)
          return (
            <span className={`inline-flex items-center gap-1 text-xs ${
              allSuccess ? 'text-emerald-400' : total === 0 ? 'text-foreground/40' : 'text-amber-400'
            }`}>
              {allSuccess ? (
                <CheckCircle size={14} />
              ) : (
                <XCircle size={14} />
              )}
              {successful}/{total}
            </span>
          )
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
          You do not have permission to view alert history.
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
            <History className="text-brand-400" />
            Alert History
          </h1>
          <p className="text-sm text-foreground/60 mt-1">
            View all sent alerts and their delivery status
          </p>
        </div>
      </div>

      {/* Summary */}
      <div className="card flex flex-wrap items-center gap-4 bg-foreground/5 border-foreground/10">
        <div className="text-sm text-foreground/60">
          {total} alert{total !== 1 ? 's' : ''} sent
        </div>
      </div>

      {/* Table */}
      <DataTable
        columns={columns}
        rows={items}
        loading={loading}
        showIndex
        emptyText="No alerts have been sent yet."
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
        title="Alert Details"
        size="lg"
      >
        {viewing && (
          <div className="space-y-4 text-sm">
            {/* Alert Info */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <span className="text-foreground/60 block mb-1">Title</span>
                <span className="font-medium">{viewing.alert_title || '—'}</span>
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Severity</span>
                {(() => {
                  const sevInfo = SEVERITY_LEVELS.find(s => s.value === viewing.severity)
                  return (
                    <span className={`px-2 py-0.5 rounded text-xs ${sevInfo?.color || 'text-foreground/60'} bg-foreground/5`}>
                      {sevInfo?.label || viewing.severity}
                    </span>
                  )
                })()}
              </div>
            </div>

            <div>
              <span className="text-foreground/60 block mb-1">Message</span>
              <div className="p-3 bg-foreground/5 rounded-lg text-foreground/80">
                {viewing.alert_message || '—'}
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <span className="text-foreground/60 block mb-1">Event Type</span>
                <span>{viewing.event_type || '—'}</span>
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Event ID</span>
                <span className="font-mono text-xs">{viewing.event_id || '—'}</span>
              </div>
            </div>

            <div>
              <span className="text-foreground/60 block mb-1">Sent At</span>
              <span>{viewing.created_at ? new Date(viewing.created_at).toLocaleString() : '—'}</span>
            </div>

            {/* Delivery Results */}
            <div>
              <span className="text-foreground/60 block mb-2">Delivery Results</span>
              <div className="space-y-2">
                {viewing.delivery_results?.length ? (
                  viewing.delivery_results.map((result, i) => (
                    <div
                      key={i}
                      className={`flex items-center justify-between p-3 rounded-lg ${
                        result.success ? 'bg-emerald-500/10' : 'bg-rose-500/10'
                      }`}
                    >
                      <div className="flex items-center gap-2">
                        {result.success ? (
                          <CheckCircle size={16} className="text-emerald-400" />
                        ) : (
                          <XCircle size={16} className="text-rose-400" />
                        )}
                        <span className="capitalize font-medium">{result.channel}</span>
                      </div>
                      <div className="text-right">
                        {result.success ? (
                          <span className="text-xs text-emerald-400">Delivered</span>
                        ) : (
                          <span className="text-xs text-rose-400">{result.error || 'Failed'}</span>
                        )}
                        {result.sent_at && (
                          <div className="text-xs text-foreground/40">
                            {new Date(result.sent_at).toLocaleTimeString()}
                          </div>
                        )}
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-foreground/40 text-center py-4">No delivery results</div>
                )}
              </div>
            </div>

            {/* IDs */}
            <div className="pt-2 border-t border-foreground/10">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
                <div>
                  <span className="text-foreground/40 block">Alert ID</span>
                  <span className="font-mono">{viewing.id}</span>
                </div>
                <div>
                  <span className="text-foreground/40 block">Config ID</span>
                  <span className="font-mono">{viewing.alert_config_id}</span>
                </div>
              </div>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
