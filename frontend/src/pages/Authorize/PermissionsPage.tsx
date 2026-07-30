import React, { useMemo, useState, useCallback } from 'react'
import { useQuery, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import toast from 'react-hot-toast'

import DataTable, { type Column } from '../../ui/DataTable'
import Modal from '../../ui/Modal'

import {
  AdmingetPermissionResources,
  setVisibilityByPrefix,
  setVisibilityByName,
  type PermissionResourcesResult,
  type PermissionResource,
  type PermissionItem,
} from '../../utils/permissions'

function toTitle(resource: string) {
  return resource.replace(/_/g, ' ')
}
function capFirst(s?: string) {
  if (!s) return ''
  return s.charAt(0).toUpperCase() + s.slice(1)
}
function getPrefixLower(r: PermissionResource): string {
  const fromResource = (r.resource || '').trim()
  if (fromResource) return fromResource.toLowerCase()
  const first = r.permissions?.[0]?.name || ''
  const derived = first.split('.')[0] || ''
  return derived.toLowerCase()
}
function currentVisibility(r: PermissionResource): 'ALL' | 'HQ_ONLY' | string {
  return (r.permissions?.[0]?.visibility as any) ?? 'ALL'
}
function oppositeVisibility(v: string): 'ALL' | 'HQ_ONLY' {
  return v === 'ALL' ? 'HQ_ONLY' : 'ALL'
}
function aggregateVisibility(r: PermissionResource): 'ALL' | 'HQ_ONLY' | 'MIXED' {
  const vals = Array.from(
    new Set((r.permissions ?? []).map(p => (p.visibility ?? 'ALL') as 'ALL' | 'HQ_ONLY'))
  )
  if (vals.length === 0) return 'ALL'
  if (vals.length === 1) return vals[0]
  return 'MIXED'
}
function badgeClass(v: 'ALL' | 'HQ_ONLY' | 'MIXED') {
  if (v === 'ALL') return 'bg-green-600/20 text-green-400'
  if (v === 'HQ_ONLY') return 'bg-red-600/20 text-red-400'
  return 'bg-amber-600/20 text-amber-400'
}

export default function PermissionsTablePage() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [rowVis, setRowVis] = useState<Record<number, 'ALL' | 'HQ_ONLY'>>({})
  const [openManage, setOpenManage] = useState(false)
  const [activeResource, setActiveResource] = useState<PermissionResource | null>(null)
  const [permVis, setPermVis] = useState<Record<number, 'ALL' | 'HQ_ONLY'>>({})
  const [savingByName, setSavingByName] = useState<Record<number, boolean>>({})

  const qc = useQueryClient()

  const params = useMemo(() => ({ page, page_size: pageSize }), [page, pageSize])
  const queryKey = useMemo(() => ['permission-resources', params] as const, [params])

  const { data, isFetching: loading } = useQuery<PermissionResourcesResult>({
    queryKey,
    queryFn: () => AdmingetPermissionResources(params),
    placeholderData: keepPreviousData,
  })

  const resources: PermissionResource[] = data?.items ?? []
  const totalResourcesAll = data?.total ?? 0
  const pageFromServer = data?.page ?? page
  const pageSizeFromServer = data?.page_size ?? pageSize

  const refresh = useCallback(
    async (msg?: string) => {
      await qc.invalidateQueries({ queryKey })
      if (msg) toast.success(msg)
    },
    [qc, queryKey]
  )

  async function applyPrefixVisibility(r: PermissionResource) {
    const agg = aggregateVisibility(r)
    if (agg === 'MIXED') {
      toast.error('This resource is MIXED. Use “Manage” to update individual permissions.')
      return
    }
    const prefix = getPrefixLower(r)
    const current = currentVisibility(r)
    const selected = rowVis[r.resource_id] ?? oppositeVisibility(current)
    if (!prefix) {
      toast.error('Could not derive prefix for this resource')
      return
    }
    try {
      await setVisibilityByPrefix({ prefix, visibility: selected })
      await refresh(`Visibility updated for prefix "${prefix}"`)
    } catch (e: any) {
      toast.error(e?.message || 'Failed to update prefix visibility')
    }
  }

  function openManageModal(resource: PermissionResource) {
    setActiveResource(resource)
    const seed: Record<number, 'ALL' | 'HQ_ONLY'> = {}
    for (const p of resource.permissions ?? []) {
      const v = (p.visibility as 'ALL' | 'HQ_ONLY') ?? 'ALL'
      seed[p.id] = v
    }
    setPermVis(seed)
    setOpenManage(true)
  }

  async function applyPermissionVisibility(p: PermissionItem) {
    const desired = permVis[p.id] ?? ((p.visibility as 'ALL' | 'HQ_ONLY') || 'ALL')
    try {
      setSavingByName(s => ({ ...s, [p.id]: true }))
      await setVisibilityByName({ name: p.name, visibility: desired })
      await refresh('Permission visibility updated')
    } catch (e: any) {
      toast.error(e?.message || 'Failed to update permission visibility')
    } finally {
      setSavingByName(s => {
        const { [p.id]: _drop, ...rest } = s
        return rest
      })
    }
  }

  const columns: Column<PermissionResource>[] = useMemo(
    () => [
      {
        key: 'resource',
        title: 'Resource',
        className: 'w-[16%] min-w-0 whitespace-normal break-words capitalize',
        sortable: true,
        render: r => <span className="font-medium">{toTitle(r.resource)}</span>,
      },
      {
        key: 'permissions',
        title: 'Permissions (name - description)',
        className: 'w-[50%] min-w-0 whitespace-normal break-words',
        render: r => {
          const agg = aggregateVisibility(r)
          const mixed = agg === 'MIXED'
          return (
            <ul className="space-y-1 text-sm">
              {(r.permissions ?? []).map(p => {
                const vis = (p.visibility as 'ALL' | 'HQ_ONLY') ?? 'ALL'
                return (
                  <li key={p.id} className="leading-snug flex items-start gap-2">
                    <span className="min-w-0">
                      {p.name ? capFirst(p.name) : '—'}
                      {p.description ? ` - ${p.description}` : ''}
                    </span>
                    {mixed && (
                      <span
                        className={`px-1.5 py-0.5 rounded-full text-[11px] font-medium shrink-0 ${badgeClass(
                          vis
                        )}`}
                      >
                        {vis}
                      </span>
                    )}
                  </li>
                )
              })}
            </ul>
          )
        },
      },
      {
        key: 'visibility_and_action',
        title: 'Visibility & Change',
        className: 'w-[28%] min-w-0',
        render: r => {
          const agg = aggregateVisibility(r) // ALL | HQ_ONLY | MIXED
          // Default opposite only for non-MIXED; for MIXED, keep a value but we'll disable controls.
          const defaultOpposite = agg === 'ALL' ? 'HQ_ONLY' : 'ALL'
          const selectValue = rowVis[r.resource_id] ?? defaultOpposite
          const disabled = agg === 'MIXED'

          return (
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-end">
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${badgeClass(agg)}`}>
                {agg}
              </span>

              <div className="flex items-center gap-2">
                <select
                  className={`input h-8 w-[120px] ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
                  value={selectValue}
                  onChange={e =>
                    setRowVis(prev => ({
                      ...prev,
                      [r.resource_id]: e.target.value as 'ALL' | 'HQ_ONLY',
                    }))
                  }
                  disabled={disabled}
                  title={disabled ? 'MIXED: change individual permissions via Manage' : 'Select visibility'}
                >
                  <option value="ALL">ALL</option>
                  <option value="HQ_ONLY">HQ_ONLY</option>
                </select>

                <button
                  className={`inline-flex h-8 items-center justify-center rounded-lg px-3 text-sm font-medium
                              border border-transparent text-foreground
                              ${disabled ? 'bg-emerald-600/40 cursor-not-allowed' : 'bg-emerald-600 hover:bg-emerald-700'}`}
                  onClick={() => applyPrefixVisibility(r)}
                  disabled={disabled}
                  title={disabled ? 'MIXED: use Manage to update individual permissions' : 'Apply visibility for this resource prefix'}
                >
                  Apply
                </button>

                <button
                  className="inline-flex h-8 items-center justify-center rounded-lg px-3 text-sm font-medium
                             bg-blue-600 hover:bg-blue-700 text-white border border-transparent"
                  onClick={() => openManageModal(r)}
                  title="Manage single permissions in this resource"
                >
                  Manage
                </button>
              </div>
            </div>
          )
        },
      },
    ],
    [rowVis]
  )

  return (
    <div className="grid gap-4 md:gap-6">
      {/* Header */}
      <div className="relative overflow-hidden rounded-2xl border border-foreground/10 bg-gradient-to-br from-indigo-600/30 via-violet-600/20 to-fuchsia-600/10 p-6">
        <h1 className="text-2xl md:text-3xl font-semibold">Permissions</h1>
        <p className="text-foreground/70 mt-1">
          Manage visibility by resource or open the modal to update individual permissions.
        </p>
      </div>

      {/* Toolbar */}
      <div className="flex items-center justify-between rounded-lg border border-foreground/10 bg-foreground/5 px-4 py-2">
        <span className="text-sm text-foreground/70">
          Total resources:{' '}
          <span className="font-semibold text-foreground">{totalResourcesAll}</span>
        </span>
        <div className="text-sm text-foreground/60">Page {pageFromServer}</div>
      </div>

      {/* Table */}
      <div className="flex justify-center">
        <div className="w-full max-w-6xl">
          <DataTable
            columns={columns}
            rows={resources}
            loading={loading}
            showIndex
            pagination={{
              page: pageFromServer,
              pageSize: pageSizeFromServer,
              total: totalResourcesAll,
              onPageChange: setPage,
              onPageSizeChange: s => {
                setPageSize(s)
                setPage(1)
              },
            }}
          />
        </div>
      </div>

      {/* Manage Modal */}
      <Modal
        open={openManage}
        onClose={() => {
          setOpenManage(false)
          setActiveResource(null)
          setPermVis({})
        }}
        title={
          activeResource
            ? `Manage: ${toTitle(activeResource.resource)}`
            : 'Manage permissions'
        }
      >
        {activeResource ? (
          <div className="grid gap-3">
            {(activeResource.permissions ?? []).map((p) => {
              const current = (p.visibility as 'ALL' | 'HQ_ONLY') ?? 'ALL'
              const value = permVis[p.id] ?? current
              return (
                <div
                  key={p.id}
                  className="flex items-start gap-3 rounded-lg border border-foreground/10 bg-foreground/5 px-3 py-2"
                >
                  <div className="min-w-0 flex-1">
                    <div className="font-medium truncate">{capFirst(p.name)}</div>
                    {p.description && (
                      <div className="text-xs text-foreground/60">{p.description}</div>
                    )}
                    <div className="mt-1">
                      <span
                        className={`px-2 py-0.5 rounded-full text-[11px] font-medium ${badgeClass(
                          current
                        )}`}
                      >
                        {current}
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <select
                      className="input h-8 w-[120px]"
                      value={value}
                      onChange={e =>
                        setPermVis(prev => ({
                          ...prev,
                          [p.id]: e.target.value as 'ALL' | 'HQ_ONLY',
                        }))
                      }
                    >
                      <option value="ALL">ALL</option>
                      <option value="HQ_ONLY">HQ_ONLY</option>
                    </select>
                    <button
                      className="inline-flex h-8 items-center justify-center rounded-lg px-3 text-sm font-medium
                                 bg-blue-600 hover:bg-blue-700 text-white border border-transparent"
                      disabled={!!savingByName[p.id]}
                      onClick={() => applyPermissionVisibility(p)}
                      title={`Apply to ${p.name}`}
                    >
                      {savingByName[p.id] ? 'Saving…' : 'Apply'}
                    </button>
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          <div className="text-foreground/70 text-sm">No resource selected.</div>
        )}
      </Modal>
    </div>
  )
}
