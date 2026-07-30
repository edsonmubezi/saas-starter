import React, { useCallback, useMemo, useState } from 'react'
import { useQuery, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import { Bell, Eye, Pencil, Plus, Trash2, CheckCircle, XCircle } from 'lucide-react'
import toast from 'react-hot-toast'
import Modal from '../../ui/Modal'
import DataTable, { type Column } from '../../ui/DataTable'
import { useAuth } from '../../state/AuthContext'

import {
  getAlertConfigs,
  getAlertConfig,
  deleteAlertConfig,
  type AlertConfigsResult,
  type AlertConfig,
  SEVERITY_LEVELS,
} from '../../utils/alerting'

import AlertConfigForm, { type AlertConfigFormValues } from './AlertConfigForm'
import { useConfirm } from '../../ui/ConfirmProvider'

export default function AlertConfigsPage() {
  const confirm = useConfirm()
  const { user } = useAuth()
  const perms: string[] = Array.isArray((user as any)?.permissions) ? (user as any).permissions : []

  // For now, use admin permissions - adjust based on your permission structure
  const canView = perms.some(p => p.includes('admin') || p.includes('alerting') || p.includes('security'))
  const canCreate = canView
  const canEdit = canView

  // Local state
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [openCreate, setOpenCreate] = useState(false)
  const [editing, setEditing] = useState<AlertConfig | null>(null)
  const [viewing, setViewing] = useState<AlertConfig | null>(null)

  const qc = useQueryClient()

  // Query params
  const params = useMemo(() => ({ page, page_size: pageSize }), [page, pageSize])
  const queryKey = useMemo(() => ['alertConfigs', params] as const, [params])

  // Query
  const { data, isFetching: loading } = useQuery<AlertConfigsResult>({
    queryKey,
    queryFn: () => getAlertConfigs(params),
    placeholderData: keepPreviousData,
    enabled: canView,
  })

  const items = data?.items ?? []
  const total = data?.total ?? 0

  // Helpers
  const refresh = useCallback(
    async (msg?: string) => {
      await qc.invalidateQueries({ queryKey })
      if (msg) toast.success(msg)
    },
    [qc, queryKey]
  )

  const handleDelete = async (row: AlertConfig) => {
    const ok = await confirm({
      title: 'Delete Alert Configuration',
      description: `Are you sure you want to delete "${row.name}"? You will no longer receive notifications for this configuration.`,
      confirmText: 'Delete',
      cancelText: 'Cancel',
      variant: 'destructive',
    })
    if (!ok) return
    try {
      await deleteAlertConfig(row.id)
      await refresh('Alert configuration deleted')
    } catch (e: any) {
      toast.error(e?.message || 'Delete failed')
    }
  }

  // Columns
  const columns: Column<AlertConfig>[] = useMemo(
    () => [
      {
        key: 'name',
        title: 'Name',
        sortable: true,
        render: (r) => (
          <div className="flex items-center gap-2">
            <Bell size={16} className="text-brand-400" />
            <span className="font-medium">{r.name || '—'}</span>
          </div>
        ),
      },
      {
        key: 'enabled',
        title: 'Status',
        render: (r) => (
          <span
            className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs ${
              r.enabled
                ? 'bg-emerald-500/20 text-emerald-400'
                : 'bg-foreground/10 text-foreground/50'
            }`}
          >
            {r.enabled ? (
              <>
                <CheckCircle size={12} /> Active
              </>
            ) : (
              <>
                <XCircle size={12} /> Disabled
              </>
            )}
          </span>
        ),
      },
      {
        key: 'severities',
        title: 'Severities',
        render: (r) => (
          <div className="flex flex-wrap gap-1">
            {r.severities?.map((sev) => {
              const sevInfo = SEVERITY_LEVELS.find((s) => s.value === sev)
              return (
                <span
                  key={sev}
                  className={`px-1.5 py-0.5 rounded text-xs ${sevInfo?.color || 'text-foreground/60'} bg-foreground/5`}
                >
                  {sevInfo?.label || sev}
                </span>
              )
            })}
          </div>
        ),
      },
      {
        key: 'channels',
        title: 'Channels',
        render: (r) => (
          <span className="text-foreground/70">
            {r.channels?.filter((c) => c.enabled).length || 0} configured
          </span>
        ),
      },
      {
        key: 'cooldown',
        title: 'Cooldown',
        render: (r) => (
          <span className="text-foreground/60 text-xs">
            {r.cooldown_seconds ? `${Math.round(r.cooldown_seconds / 60)}m` : '—'}
          </span>
        ),
      },
      {
        key: 'created_at',
        title: 'Created',
        sortable: true,
        render: (r) => (
          <span className="text-foreground/60 text-xs">
            {r.created_at ? new Date(r.created_at).toLocaleDateString() : '—'}
          </span>
        ),
      },
      {
        key: 'actions',
        title: 'Actions',
        className: 'text-right w-[140px]',
        render: (o) => (
          <div className="flex h-10 items-center justify-end gap-1">
            {canView && (
              <button
                className="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-foreground/10"
                title="View"
                onClick={async () => {
                  try {
                    const fresh = await getAlertConfig(o.id)
                    setViewing(fresh)
                  } catch {
                    setViewing(o)
                  }
                }}
              >
                <Eye size={16} />
              </button>
            )}
            {canEdit && (
              <button
                className="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-foreground/10"
                title="Edit"
                onClick={() => setEditing(o)}
              >
                <Pencil size={16} />
              </button>
            )}
            <button
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-foreground/10 text-rose-300"
              title="Delete"
              onClick={() => handleDelete(o)}
            >
              <Trash2 size={16} />
            </button>
          </div>
        ),
      },
    ],
    [canView, canEdit]
  )

  // No permission UI
  if (!canView) {
    return (
      <div className="card p-6 bg-foreground/5 border-foreground/10">
        <div className="text-sm text-foreground/80 font-medium">
          You do not have permission to view alert configurations.
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
            <Bell className="text-brand-400" />
            Alert Configurations
          </h1>
          <p className="text-sm text-foreground/60 mt-1">
            Configure alerts for security events and system notifications
          </p>
        </div>
      </div>

      {/* Toolbar */}
      <div className="card flex flex-wrap items-center justify-between gap-3 bg-foreground/5 border-foreground/10">
        <div className="text-sm text-foreground/60">
          {total} configuration{total !== 1 ? 's' : ''}
        </div>
        {canCreate && (
          <button className="btn flex items-center gap-2" onClick={() => setOpenCreate(true)}>
            <Plus size={16} /> New Alert Config
          </button>
        )}
      </div>

      {/* Table */}
      <DataTable
        columns={columns}
        rows={items}
        loading={loading}
        showIndex
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

      {/* Create modal */}
      <Modal
        open={openCreate}
        onClose={() => setOpenCreate(false)}
        title="Create Alert Configuration"
        size="lg"
      >
        {canCreate ? (
          <AlertConfigForm
            mode="create"
            onSubmit={async () => {
              setOpenCreate(false)
              setPage(1)
              await refresh('Alert configuration created')
            }}
            onCancel={() => setOpenCreate(false)}
          />
        ) : (
          <div className="p-2 text-sm text-foreground/70">
            You do not have permission to create alert configurations.
          </div>
        )}
      </Modal>

      {/* Edit modal */}
      <Modal
        open={!!editing}
        onClose={() => setEditing(null)}
        title="Edit Alert Configuration"
        size="lg"
      >
        {editing &&
          (canEdit ? (
            <AlertConfigForm
              mode="edit"
              initial={editing}
              onSubmit={async () => {
                setEditing(null)
                await refresh('Alert configuration updated')
              }}
              onCancel={() => setEditing(null)}
            />
          ) : (
            <div className="p-2 text-sm text-foreground/70">
              You do not have permission to edit alert configurations.
            </div>
          ))}
      </Modal>

      {/* View modal */}
      <Modal
        open={!!viewing}
        onClose={() => setViewing(null)}
        title="Alert Configuration Details"
        size="lg"
      >
        {viewing && (
          <div className="space-y-4 text-sm">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <span className="text-foreground/60 block mb-1">Name</span>
                <span className="font-medium">{viewing.name || '—'}</span>
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Status</span>
                <span
                  className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs ${
                    viewing.enabled
                      ? 'bg-emerald-500/20 text-emerald-400'
                      : 'bg-foreground/10 text-foreground/50'
                  }`}
                >
                  {viewing.enabled ? 'Active' : 'Disabled'}
                </span>
              </div>
            </div>

            <div>
              <span className="text-foreground/60 block mb-1">Severity Levels</span>
              <div className="flex flex-wrap gap-1">
                {viewing.severities?.map((sev) => {
                  const sevInfo = SEVERITY_LEVELS.find((s) => s.value === sev)
                  return (
                    <span
                      key={sev}
                      className={`px-2 py-1 rounded text-xs ${sevInfo?.color || 'text-foreground/60'} bg-foreground/5`}
                    >
                      {sevInfo?.label || sev}
                    </span>
                  )
                })}
              </div>
            </div>

            <div>
              <span className="text-foreground/60 block mb-1">Event Types</span>
              <div className="flex flex-wrap gap-1">
                {viewing.event_types?.length ? (
                  viewing.event_types.map((et) => (
                    <span key={et} className="px-2 py-1 rounded text-xs bg-foreground/5">
                      {et}
                    </span>
                  ))
                ) : (
                  <span className="text-foreground/40">All events</span>
                )}
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <div>
                <span className="text-foreground/60 block mb-1">Cooldown</span>
                <span>{Math.round(viewing.cooldown_seconds / 60)} minutes</span>
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Threshold</span>
                <span>{viewing.threshold_count} events</span>
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Window</span>
                <span>{Math.round(viewing.window_seconds / 60)} minutes</span>
              </div>
            </div>

            <div>
              <span className="text-foreground/60 block mb-1">Channels</span>
              <div className="space-y-2">
                {viewing.channels?.map((ch, i) => (
                  <div
                    key={i}
                    className={`flex items-center gap-2 px-3 py-2 rounded-lg ${
                      ch.enabled ? 'bg-foreground/5' : 'bg-foreground/5 opacity-50'
                    }`}
                  >
                    <span className="capitalize font-medium">{ch.channel_type}</span>
                    <span
                      className={`text-xs px-1.5 py-0.5 rounded ${
                        ch.enabled ? 'bg-emerald-500/20 text-emerald-400' : 'bg-foreground/10 text-foreground/50'
                      }`}
                    >
                      {ch.enabled ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-2 border-foreground/10">
              <div>
                <span className="text-foreground/60 block mb-1">Created</span>
                <span className="text-xs">
                  {viewing.created_at ? new Date(viewing.created_at).toLocaleString() : '—'}
                </span>
              </div>
              <div>
                <span className="text-foreground/60 block mb-1">Last Updated</span>
                <span className="text-xs">
                  {viewing.updated_at ? new Date(viewing.updated_at).toLocaleString() : '—'}
                </span>
              </div>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
