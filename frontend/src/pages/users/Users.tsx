// src/pages/users/UsersPage.tsx
import React, { useCallback, useMemo, useState } from 'react'
import { useQuery, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import { Eye, Pencil, Plus, Trash2, Users, CheckCircle, XCircle } from 'lucide-react'
import toast from 'react-hot-toast'

import {
  getUsers,
  createUser,
  updateUser,
  deleteUser,
  activateUser,
  deactivateUser,
  sendPasswordReset,
  type UserListItem,
  type UsersResult,
  type CreateUserInput,
} from '../../utils/users'

import Modal from '../../ui/Modal'
import ConfirmModal from '../../ui/ConfirmModal'
import DataTable, { type Column } from '../../ui/DataTable'
import UserForm, { type UserFormValues } from './UserForm'
import OrganizationFilter from '../../ui/OrganizationFilter'
import StatusBadge from '../../ui/StatusBadge'
import { StatsGrid } from '../../ui/StatsCard'
import useDebounce from '../../state/useDebounce'

// 🔸 bring in auth/permissions
import { useAuth } from '../../state/AuthContext'

// normalize server status to 'active' | 'disabled'
const toStatus = (s: UserListItem['status']): 'active' | 'disabled' =>
  typeof s === 'string'
    ? (s === 'disabled' ? 'disabled' : 'active')
    : (s === false ? 'disabled' : 'active')

export default function UsersPage() {
  const { user } = useAuth()
  const perms: string[] = Array.isArray((user as any)?.permissions) ? (user as any).permissions : []
  const canSeeOrg = perms.includes('admin.organization.view') // 👈 single source of truth
  const canCreateUser = perms.includes('admin.user.create')
  const canEditUser = perms.includes('admin.user.edit')
  const canDeleteUser = perms.includes('admin.user.delete')
  const canEditStatus = perms.includes('admin.user.edit_status')

  // ── Local state ──────────────────────────────────────────────────────────────
  const [search, setSearch] = useState('')
  const debounced = useDebounce(search, 300)

  const [orgId, setOrgId] = useState<string>('') // '' means all
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)

  const [openCreate, setOpenCreate] = useState(false)
  const [editing, setEditing] = useState<UserListItem | null>(null)
  const [deleting, setDeleting] = useState<UserListItem | null>(null)
  const [viewing, setViewing] = useState<UserListItem | null>(null)

  const [toggling, setToggling] = useState(false)
  const [resetting, setResetting] = useState(false)

  const qc = useQueryClient()

  // ── Query params / key ───────────────────────────────────────────────────────
  const params = useMemo(
    () => ({
      q: debounced || undefined,
      page,
      page_size: pageSize,
      // only send org filter if user can see orgs
      organization_id: canSeeOrg ? (orgId || undefined) : undefined,
    }),
    [debounced, page, pageSize, orgId, canSeeOrg],
  )

  const queryKey = useMemo(() => ['users', params] as const, [params])

  // ── Query ────────────────────────────────────────────────────────────────────
  const { data, isFetching: loading } = useQuery<UsersResult>({
    queryKey,
    queryFn: () => getUsers(params),
    placeholderData: keepPreviousData,
  })

  const items = data?.items ?? []
  const total = data?.total ?? 0
  const pageFromServer = data?.page ?? page
  const pageSizeFromServer = data?.page_size ?? pageSize

  // Calculate statistics
  const stats = useMemo(() => {
    const active = items.filter(u => toStatus(u.status) === 'active').length
    const disabled = items.filter(u => toStatus(u.status) === 'disabled').length
    return { total, active, disabled }
  }, [items, total])

  // ── Helpers ──────────────────────────────────────────────────────────────────
  const refresh = useCallback(
    async (msg?: string) => {
      await qc.invalidateQueries({ queryKey })
      if (msg) toast.success(msg)
    },
    [qc, queryKey],
  )

  const handleCreate = async (values: UserFormValues) => {
    const payload: CreateUserInput = {
      fullname: values.fullname.trim(),
      email: values.email.trim(),
      password: values.password ?? '',
      status: values.status,
      role_id: values.role_id,
      // do not send organization_id if user can't see orgs
      ...(canSeeOrg ? { organization_id: values.organization_id } : {}),
    }

    await createUser(payload)
    setOpenCreate(false)
    setPage(1)
    await refresh('User created')
  }

  const handleEditSubmit = async (values: UserFormValues) => {
    if (!editing) return
    // similarly, avoid sending org when not allowed
    const safeValues: Partial<UserFormValues> = {
      fullname: values.fullname,
      email: values.email,
      status: values.status,
      role_id: values.role_id,
      ...(canSeeOrg ? { organization_id: values.organization_id } : {}),
    }
    await updateUser(editing.id, safeValues as UserFormValues)
    setEditing(null)
    await refresh('User updated')
  }

  const handleDeleteConfirm = async () => {
    if (!deleting) return
    await deleteUser(deleting.id)
    setDeleting(null)
    await refresh('User deleted')
  }

  const handleToggleStatus = async (u: UserListItem) => {
    try {
      setToggling(true)
      const isActive = toStatus(u.status) === 'active'
      if (isActive) {
        await deactivateUser(u.id)
      } else {
        await activateUser(u.id)
      }
      setViewing(null)
      await refresh(isActive ? 'User deactivated' : 'User activated')
    } catch (e: any) {
      toast.error(e?.message || 'Failed to update status')
    } finally {
      setToggling(false)
    }
  }

  const handleResetPassword = async (u: UserListItem) => {
    try {
      setResetting(true)
      await sendPasswordReset(u.id)
      toast.success('Password reset link sent')
    } catch (e: any) {
      toast.error(e?.message || 'Failed to send reset link')
    } finally {
      setResetting(false)
    }
  }

  // ── Columns (conditionally include Organization) ─────────────────────────────
  const columns: Column<UserListItem & { actions?: React.ReactNode }>[] = useMemo(() => {
    const base: Column<UserListItem & { actions?: React.ReactNode }>[] = [
      {
        key: 'fullname',
        title: 'Full name',
        sortable: true,
        render: r => r.fullname || '—',
      },
      {
        key: 'email',
        title: 'Email',
        sortable: true,
        render: r => r.email,
      },
      {
        key: 'role',
        title: 'Role',
        sortable: true,
        render: r => r.role || '—',
      },
      {
        key: 'status',
        title: 'Status',
        sortable: true,
        render: r => (
          <div className="flex h-10 items-center">
            <StatusBadge status={toStatus(r.status)} type="user" />
          </div>
        ),
      },
    ]

    if (canSeeOrg) {
      base.splice(3, 0, {
        key: 'organization',
        title: 'Organization',
        sortable: true,
        render: r => (r as any).organization || '—',
      })
    }

    base.push({
      key: 'actions',
      title: 'Actions',
      className: 'text-right w-[160px]',
      render: u => (
        <div className="flex h-10 items-center justify-end gap-1">
          <button
            className="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-foreground/10"
            title="View"
            onClick={() => setViewing(u)}
          >
            <Eye size={16} />
          </button>
          {canEditUser && (
            <button
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-foreground/10"
              title="Edit"
              onClick={() => setEditing(u)}
            >
              <Pencil size={16} />
            </button>
          )}
          {canDeleteUser && (
            <button
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-foreground/10"
              title="Delete"
              onClick={() => setDeleting(u)}
            >
              <Trash2 size={16} />
            </button>
          )}
        </div>
      ),
    })

    return base
  }, [canSeeOrg, canEditUser, canDeleteUser])

  // ── Render ───────────────────────────────────────────────────────────────────
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-xl md:text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            User Management
          </h1>
          <p className="text-foreground/60 text-sm mt-1">Manage system users and their access</p>
        </div>
        <div className="flex gap-3 flex-wrap">
          {canCreateUser && (
            <button
              className="px-4 py-2.5 rounded-lg bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-medium shadow-lg shadow-blue-500/20 transition-all flex items-center gap-2"
              onClick={() => setOpenCreate(true)}
            >
              <Plus className="w-4 h-4" />
              Add User
            </button>
          )}
        </div>
      </div>

      {/* Statistics Cards */}
      <StatsGrid
        columns={3}
        stats={[
          { label: 'Total Users', value: stats.total, icon: <Users className="w-6 h-6" />, color: 'blue' },
          { label: 'Active', value: stats.active, icon: <CheckCircle className="w-6 h-6" />, color: 'green' },
          { label: 'Disabled', value: stats.disabled, icon: <XCircle className="w-6 h-6" />, color: 'red' },
        ]}
      />

      {/* Toolbar */}
      <div className="bg-foreground/5 border border-foreground/10 rounded-xl p-4 backdrop-blur">
        <div className="flex flex-wrap items-center gap-3">
          <input
            className="flex-1 min-w-[200px] max-w-md bg-foreground/5 border border-foreground/10 rounded-lg px-4 py-2.5 text-foreground placeholder:text-foreground/40 outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500/50 transition-all"
            placeholder="Search users..."
            value={search}
            onChange={e => {
              setSearch(e.target.value)
              setPage(1)
            }}
          />
          {/* hide org filter if no permission */}
          {canSeeOrg && (
            <OrganizationFilter
              value={orgId}
              onChange={v => {
                setOrgId(v)
                setPage(1)
              }}
            />
          )}
        </div>
      </div>

      {/* Table */}
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

      {/* Create modal */}
      <Modal open={openCreate} onClose={() => setOpenCreate(false)} title="Create user">
        <UserForm
          mode="create"
          // if user can't see orgs, the form can still render (your component can hide its org field by the same flag)
          onSubmit={async v => {
            try {
              await handleCreate(v)
            } catch (e: any) {
              toast.error(e?.message || 'Failed to create')
              throw e
            }
          }}
        />
      </Modal>

      {/* Edit modal */}
      <Modal open={!!editing} onClose={() => setEditing(null)} title="Edit user">
        {editing && (
          <UserForm
            mode="edit"
            initial={{
              fullname: editing.fullname || '',
              email: editing.email,
              status: toStatus(editing.status),
              role_id: (editing as any).role_id,
              organization_id: (editing as any).organization_id,
            }}
            onSubmit={async v => {
              try {
                await handleEditSubmit(v)
              } catch (e: any) {
                toast.error(e?.message || 'Failed to update')
                throw e
              }
            }}
          />
        )}
      </Modal>

      {/* Details modal */}
      <Modal open={!!viewing} onClose={() => setViewing(null)} title="User details">
        {viewing && (
          <div className="space-y-3">
            <div className="text-sm">
              <span className="text-foreground/60">Full name:</span>{' '}
              <span className="font-medium">{viewing.fullname || '—'}</span>
            </div>

            <div className="text-sm">
              <span className="text-foreground/60">Email:</span>{' '}
              <span className="font-medium break-all">{viewing.email || '—'}</span>
            </div>

            <div className="text-sm">
              <span className="text-foreground/60">Role:</span>{' '}
              <span className="font-medium">{viewing.role || '—'}</span>
            </div>

            {/* hide organization row if no permission */}
            {canSeeOrg && (
              <div className="text-sm">
                <span className="text-foreground/60">Organization:</span>{' '}
                <span className="font-medium">{viewing.organization || '—'}</span>
              </div>
            )}

            {/* Actions */}
            <div className="flex flex-wrap items-center justify-end gap-2 pt-2">
              {canEditUser && (
                <button
                  className="h-8 px-3 text-xs rounded-lg bg-orange-500 hover:bg-orange-400 text-foreground disabled:opacity-50 disabled:cursor-not-allowed"
                  disabled={resetting}
                  onClick={() => handleResetPassword(viewing)}
                >
                  {resetting ? 'Sending…' : 'Reset password'}
                </button>
              )}

              {canEditStatus && (() => {
                const isActive = toStatus(viewing.status) === 'active'
                const base = 'h-8 px-3 text-xs rounded-lg text-foreground disabled:opacity-50 disabled:cursor-not-allowed'
                const color = isActive ? 'bg-green-600 hover:bg-green-500' : 'bg-red-600 hover:bg-red-500'
                return (
                  <button
                    className={`${base} ${color}`}
                    disabled={toggling}
                    onClick={() => handleToggleStatus(viewing)}
                  >
                    {toggling ? 'Working…' : isActive ? 'Deactivate' : 'Activate'}
                  </button>
                )
              })()}
            </div>
          </div>
        )}
      </Modal>

      {/* Delete modal */}
      <ConfirmModal
        open={!!deleting}
        onCancel={() => setDeleting(null)}
        onConfirm={async () => {
          try {
            await handleDeleteConfirm()
          } catch (e: any) {
            toast.error(e?.message || 'Failed to delete')
          }
        }}
        title="Delete user"
        message={deleting ? <>Are you sure you want to delete <b>{deleting.email}</b>?</> : ''}
        confirmText="Delete"
        danger
      />
    </div>
  )
}
