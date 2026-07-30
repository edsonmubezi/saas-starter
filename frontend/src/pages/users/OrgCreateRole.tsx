import React, { useCallback, useMemo, useState } from 'react'
import { useQuery, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { Building2, Plus, Pencil, RotateCw } from 'lucide-react'

import {
  getRolesOrgsPaged,
  createRole,
  updateRole,
  type RoleListItem,
  type RolesResult,
  type GetRolesParams,
} from '../../utils/roles'

import { useAuth } from '../../state/AuthContext'
import Modal from '../../ui/Modal'
import DataTable, { type Column } from '../../ui/DataTable'
import RoleForm, { type RoleFormValues } from '../Authorize/RoleForm'

export default function OrgCreateRole() {
  const { user } = useAuth()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [openCreate, setOpenCreate] = useState(false)
  const [editing, setEditing] = useState<RoleListItem | null>(null)

  // Get current user's organization ID for filtering
  const currentUserOrgId = (user as any)?.organization?.id || (user as any)?.organization_id

  // Query params
  const params: GetRolesParams = useMemo(
    () => ({
      page,
      page_size: pageSize,
      // Add organization filter if the API supports it
      organization_id: currentUserOrgId,
    }),
    [page, pageSize, currentUserOrgId]
  )

  const queryKey = useMemo(() => ['org-roles', params] as const, [params])

  // Roles query
  const { data, isFetching: loading } = useQuery<RolesResult>({
    queryKey,
    queryFn: () => getRolesOrgsPaged(params),
    placeholderData: keepPreviousData,
  })

  const items = data?.items ?? []
  const total = data?.total ?? 0
  const pageFromServer = data?.page ?? page
  const pageSizeFromServer = data?.page_size ?? pageSize

  // Helpers
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
    await refresh('Role created successfully')
  }

  const handleEditSubmit = async (values: RoleFormValues) => {
    if (!editing) return
    await updateRole(editing.id, { name: values.name.trim() })
    setEditing(null)
    await refresh('Role updated successfully')
  }

  // Columns
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

  return (
    <div className="grid gap-6">
      {/* Header */}
      <div className="flex items-center gap-3 mb-2">
        <div className="p-2 bg-blue-600/20 rounded-lg ring-1 ring-blue-500/30">
          <Building2 className="w-5 h-5 text-blue-400" />
        </div>
        <div>
          <h1 className="text-xl font-semibold text-foreground">Organization Roles</h1>
          <p className="text-sm text-foreground/60">
            Manage roles in your organization: {(user as any)?.organization?.name || 'Unknown Organization'}
          </p>
        </div>
      </div>

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
            className="inline-flex items-center gap-2 rounded-md bg-blue-500 px-3 py-1.5 text-sm font-medium text-foreground hover:bg-blue-600 transition"
            onClick={() => setOpenCreate(true)}
          >
            <Plus size={14} /> Add role
          </button>
        </div>
      </div>

      {/* Table */}
      {items.length === 0 && !loading ? (
        <div className="mx-auto w-full max-w-2xl">
          <div className="rounded-2xl border border-dashed border-foreground/15 bg-foreground/5 p-10 text-center">
            <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-2xl bg-foreground/10">
              <Plus size={20} />
            </div>
            <h3 className="text-lg font-semibold">No roles yet</h3>
            <p className="text-foreground/70 mt-1">
              Create your first role in your organization.
            </p>
            <button
              className="btn mt-4 bg-blue-500/70 hover:bg-blue-500 transition"
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
    </div>
  )
}