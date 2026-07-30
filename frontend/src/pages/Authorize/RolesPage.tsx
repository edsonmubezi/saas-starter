import React, { useCallback, useMemo, useState } from 'react'
import { useQuery, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import { Eye, Pencil, Plus, RotateCw } from 'lucide-react'
import toast from 'react-hot-toast'

import {
  getRolesPaged,
  updateRole,
  createRole,
  getRoleDetails,
  type RolesResult,
  type RoleListItem,
  type GetRolesParams,
  type RoleDetails,
  type RolePermission,
} from '../../utils/roles'

import DataTable, { type Column } from '../../ui/DataTable'
import Modal from '../../ui/Modal'
import RoleForm, { type RoleFormValues } from './RoleForm'

export default function RolesPage() {
  // ── Local state ──────────────────────────────────────────────────────────────
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [openCreate, setOpenCreate] = useState(false)
  const [editing, setEditing] = useState<RoleListItem | null>(null)

  // NEW: permissions modal state
  const [viewingRole, setViewingRole] = useState<RoleListItem | null>(null)
  const [permSearch, setPermSearch] = useState('')

  const qc = useQueryClient()

  // ── Query params / key ───────────────────────────────────────────────────────
  const params: GetRolesParams = useMemo(
    () => ({
      page,
      page_size: pageSize,
    }),
    [page, pageSize]
  )

  const queryKey = useMemo(() => ['roles', params] as const, [params])

  // ── Role list query ──────────────────────────────────────────────────────────
  const { data, isFetching: loading } = useQuery<RolesResult>({
    queryKey,
    queryFn: () => getRolesPaged(params),
    placeholderData: keepPreviousData,
  })

  const items = data?.items ?? []
  const total = data?.total ?? 0
  const pageFromServer = data?.page ?? page
  const pageSizeFromServer = data?.page_size ?? pageSize

  // ── Role details query (enabled only when viewingRole is set) ────────────────
  const {
    data: details,
    isFetching: detailsLoading,
    error: detailsError,
  } = useQuery<RoleDetails>({
    queryKey: ['role-details', viewingRole?.id],
    queryFn: () => getRoleDetails(viewingRole!.id),
    enabled: !!viewingRole?.id,
    staleTime: 60_000,
  })

  // ── Helpers ──────────────────────────────────────────────────────────────────
  const refresh = useCallback(
    async (msg?: string) => {
      await qc.invalidateQueries({ queryKey })
      if (msg) toast.success(msg)
    },
    [qc, queryKey]
  )

  const handleCreate = async (values: RoleFormValues) => {
    await createRole({ name: values.name.trim() })
    setOpenCreate(false)
    setPage(1)
    await refresh('Role created')
  }

  const handleEditSubmit = async (values: RoleFormValues) => {
    if (!editing) return
    await updateRole(editing.id, { name: values.name.trim() })
    setEditing(null)
    await refresh('Role updated')
  }

  // ── Columns ──────────────────────────────────────────────────────────────────
  const columns: Column<RoleListItem & { actions?: React.ReactNode }>[] = useMemo(
    () => [
      {
        key: 'name',
        title: 'Role',
        sortable: true,
        className: 'w-[65%]',
        render: r => (
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center rounded-full bg-foreground/10 px-2 py-1 text-xs text-foreground/80">
              {r.name?.slice(0, 1).toUpperCase()}
            </span>
            <span className="font-medium">{r.name || '—'}</span>
          </div>
        ),
      },
      {
        key: 'actions',
        title: 'Actions',
        className: 'text-right w-[35%]',
        render: r => (
          <div className="flex h-10 items-center justify-end gap-1">
            {/* NEW: View permissions button */}
            <button
              className="inline-flex h-8 items-center gap-1 rounded-lg px-2 hover:bg-foreground/10 transition"
              title="View permissions"
              onClick={() => {
                setViewingRole(r)
                setPermSearch('')
              }}
            >
              <Eye size={16} />
              <span className="text-sm">Permissions</span>
            </button>

            {/* Existing: Edit button */}
            <button
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-foreground/10 transition"
              title="Edit"
              onClick={() => setEditing(r)}
            >
              <Pencil size={16} />
            </button>
          </div>
        ),
      },
    ],
    []
  )

  // ── Permission grouping + filtering for modal ────────────────────────────────
  const grouped = useMemo(() => {
    const list = details?.permissions ?? []
    const q = permSearch.trim().toLowerCase()
    const filtered = q
      ? list.filter(p =>
          `${p.name} ${p.description ?? ''}`.toLowerCase().includes(q)
        )
      : list

    // group by resource (prefix before ".")
    const map = new Map<string, RolePermission[]>()
    for (const p of filtered) {
      const prefix = p.name.includes('.') ? p.name.split('.')[0] : p.name
      if (!map.has(prefix)) map.set(prefix, [])
      map.get(prefix)!.push(p)
    }
    // sort groups alphabetically, and permissions inside groups
    return Array.from(map.entries())
      .sort((a, b) => a[0].localeCompare(b[0]))
      .map(([k, arr]) => [k, arr.sort((x, y) => x.name.localeCompare(y.name))] as const)
  }, [details?.permissions, permSearch])

  // ── Render ───────────────────────────────────────────────────────────────────
  return (
    <div className="grid gap-6">
      {/* Toolbar card */}
      <div className="flex items-center justify-between rounded-lg border border-foreground/10 bg-foreground/5 px-4 py-2">
        <span className="text-sm text-foreground/70">
          Total roles: <span className="font-semibold text-foreground">{total}</span>
        </span>
        <div className="flex items-center gap-2">
          <button
            className="inline-flex items-center justify-center rounded-md bg-foreground/10 px-3 py-1.5 text-sm hover:bg-foreground/20 transition"
            onClick={() => refresh('Refreshed')}
            aria-label="Refresh"
          >
            <RotateCw size={14} className={loading ? 'animate-spin' : ''} />
          </button>
          <button
            className="inline-flex items-center gap-2 rounded-md bg-indigo-500 px-3 py-1.5 text-sm font-medium text-foreground hover:bg-indigo-600 transition"
            onClick={() => setOpenCreate(true)}
          >
            <Plus size={14} /> Add role
          </button>
        </div>
      </div>

      {/* Table — centered & constrained */}
      {items.length === 0 && !loading ? (
        <div className="mx-auto w-full max-w-2xl">
          <div className="rounded-2xl border border-dashed border-foreground/15 bg-foreground/5 p-10 text-center">
            <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-2xl bg-foreground/10">
              <Plus size={20} />
            </div>
            <h3 className="text-lg font-semibold">No roles yet</h3>
            <p className="text-foreground/70 mt-1">
              Create your first role to start assigning permissions to users.
            </p>
            <button
              className="btn mt-4 bg-indigo-500/70 hover:bg-indigo-500 transition"
              onClick={() => setOpenCreate(true)}
            >
              Add role
            </button>
          </div>
        </div>
      ) : (
        <div className="flex justify-center">
          <div className="w-full max-w-2xl">
            <DataTable
              columns={columns}
              rows={Array.isArray(items) ? items : []}
              loading={loading}
              showIndex
              pagination={{
                page: pageFromServer,
                pageSize: pageSizeFromServer,
                total,
                onPageChange: setPage,
                onPageSizeChange: s => {
                  setPageSize(s)
                  setPage(1)
                },
              }}
            />
          </div>
        </div>
      )}

      {/* Create modal */}
      <Modal open={openCreate} onClose={() => setOpenCreate(false)} title="Add role">
        <RoleForm
          mode="create"
          onSubmit={async v => {
            try {
              await handleCreate(v)
            } catch (e: any) {
              toast.error(e?.message || 'Failed to create role')
              throw e
            }
          }}
        />
      </Modal>

      {/* Edit modal */}
      <Modal open={!!editing} onClose={() => setEditing(null)} title="Edit role">
        {editing && (
          <RoleForm
            mode="edit"
            initial={{ name: editing.name }}
            onSubmit={async v => {
              try {
                await handleEditSubmit(v)
              } catch (e: any) {
                toast.error(e?.message || 'Failed to update role')
                throw e
              }
            }}
          />
        )}
      </Modal>

      {/* NEW: View permissions modal */}
      <Modal
        open={!!viewingRole}
        onClose={() => setViewingRole(null)}
        title={viewingRole ? `Permissions — ${viewingRole.name}` : 'Permissions'}
      >
        <div className="space-y-4">
          {/* Search/filter */}
          <div className="flex items-center gap-2">
            <input
              value={permSearch}
              onChange={e => setPermSearch(e.target.value)}
              placeholder="Filter by name or description…"
              className="w-full rounded-md border border-foreground/10 bg-foreground/5 px-3 py-2 text-sm outline-none placeholder:text-foreground/50"
            />
            <span className="text-xs text-foreground/60">
              {details?.permissions?.length ?? 0} total
            </span>
          </div>

          {/* Content */}
          {detailsError ? (
            <div className="rounded-md border border-red-500/30 bg-red-500/10 p-3 text-sm text-red-200">
              Failed to load permissions
            </div>
          ) : detailsLoading ? (
            <div className="text-sm text-foreground/70">Loading permissions…</div>
          ) : (details?.permissions?.length ?? 0) === 0 ? (
            <div className="text-sm text-foreground/70">No permissions assigned.</div>
          ) : (
            <div className="max-h-[60vh] overflow-auto pr-1">
              {grouped.map(([group, perms]) => (
                <div key={group} className="mb-4">
                  <div className="sticky top-0 z-10 mb-2 rounded-md bg-foreground/5 px-2 py-1 text-xs font-semibold uppercase tracking-wide text-foreground/70">
                    {group} <span className="text-foreground/40">({perms.length})</span>
                  </div>
                  <ul className="space-y-1">
                    {perms.map(p => (
                      <li
                        key={p.id}
                        className="rounded-md border border-foreground/10 bg-foreground/5 px-3 py-2"
                      >
                        <div className="text-sm font-medium">{p.name}</div>
                        {p.description ? (
                          <div className="text-xs text-foreground/70">{p.description}</div>
                        ) : null}
                      </li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
          )}
        </div>
      </Modal>
    </div>
  )
}
