// src/pages/users/SuperAdminUsers.tsx
import React, { useState, useMemo } from 'react'
import { useQuery, useQueryClient, keepPreviousData, useMutation } from '@tanstack/react-query'
import {
  Eye, Pencil, Plus, Trash2, Lock, Unlock, Shield, Users, Building2,
  UserCheck, Search, Mail, Power,
} from 'lucide-react'
import toast from 'react-hot-toast'
import clsx from 'clsx'

import {
  getUsers,
  getUserCounts,
  createUser,
  updateUser,
  deleteUser,
  unlockUserAccount,
  lockUserAccount,
  activateUser,
  deactivateUser,
  sendPasswordReset,
  type UserListItem,
  type CreateUserInput,
  type UserCounts,
} from '../../utils/users'

import Modal from '../../ui/Modal'
import ConfirmModal from '../../ui/ConfirmModal'
import DataTable, { type Column } from '../../ui/DataTable'
import UserForm, { type UserFormValues } from './UserForm'
import OrganizationFilter from '../../ui/OrganizationFilter'
import StatusBadge from '../../ui/StatusBadge'
import useDebounce from '../../state/useDebounce'
import { useAuth } from '../../state/AuthContext'

type TabKey = 'system' | 'org_admin'

const TAB_CONFIG: { key: TabKey; label: string; icon: React.ReactNode }[] = [
  { key: 'system',    label: 'System Users', icon: <Shield size={16} /> },
  { key: 'org_admin', label: 'Org Admins',   icon: <Building2 size={16} /> },
]

const toStatus = (s: UserListItem['status']): string =>
  typeof s === 'string' ? (s === 'disabled' ? 'disabled' : 'active') : (s === false ? 'disabled' : 'active')

export default function SuperAdminUsersPage() {
  const { user } = useAuth()
  const qc = useQueryClient()
  const perms: string[] = Array.isArray((user as any)?.permissions) ? (user as any).permissions : []

  const canCreateUser = perms.includes('admin.user.create')
  const canEditUser   = perms.includes('admin.user.edit')
  const canDeleteUser = perms.includes('admin.user.delete')
  const canUnlock     = perms.includes('admin.user.unlock')
  const canEditStatus = perms.includes('admin.user.edit_status')
  const canResetPwd   = perms.includes('admin.user.reset_password')

  // Tab state
  const [activeTab, setActiveTab] = useState<TabKey>('system')

  // Per-tab pagination
  const [pages, setPages] = useState({ system: 1, org_admin: 1 })
  const [pageSize, setPageSize] = useState(20)
  const setPage = (tab: TabKey, p: number) => setPages(prev => ({ ...prev, [tab]: p }))

  // Shared search
  const [search, setSearch] = useState('')
  const debounced = useDebounce(search, 300)

  // Org filter for tabs 2 & 3
  const [orgFilter, setOrgFilter] = useState({ org_admin: '' })

  // Per-tab sort
  const [sorts, setSorts] = useState<Record<TabKey, { by: string; dir: 'asc' | 'desc' }>>({
    system:    { by: 'full_name', dir: 'asc' },
    org_admin: { by: 'full_name', dir: 'asc' },
  })

  // Modals
  const [creating, setCreating] = useState(false)
  const [editing, setEditing]   = useState<UserListItem | null>(null)
  const [viewing, setViewing]   = useState<UserListItem | null>(null)
  const [deleting, setDeleting] = useState<UserListItem | null>(null)

  // Counts query
  const countsQuery = useQuery({
    queryKey: ['user-counts'],
    queryFn: getUserCounts,
    staleTime: 30_000,
  })
  const counts: UserCounts = countsQuery.data ?? { system: 0, org_admin: 0 }

  // Data queries per tab
  const buildQuery = (tab: TabKey) => ({
    queryKey: ['users', tab, { page: pages[tab], search: debounced, sort: sorts[tab], org: tab !== 'system' ? orgFilter[tab as 'org_admin'] : '' }],
    queryFn: () => getUsers({
      page: pages[tab],
      page_size: pageSize,
      q: debounced,
      user_type: tab,
      sort_by: sorts[tab].by,
      sort_dir: sorts[tab].dir,
      organization_id: tab !== 'system' ? (orgFilter[tab as 'org_admin'] || undefined) : undefined,
    }),
    enabled: activeTab === tab,
    placeholderData: keepPreviousData,
  })

  const systemQ    = useQuery(buildQuery('system'))
  const orgAdminQ  = useQuery(buildQuery('org_admin'))

  const activeQuery = activeTab === 'system' ? systemQ : orgAdminQ
  const data = activeQuery.data

  // Mutations
  const invalidateAll = () => {
    qc.invalidateQueries({ queryKey: ['users'] })
    qc.invalidateQueries({ queryKey: ['user-counts'] })
  }

  const createMut = useMutation({
    mutationFn: (v: CreateUserInput) => createUser(v),
    onSuccess: () => { toast.success('User created — credentials sent via email'); setCreating(false); invalidateAll() },
    onError: (e: Error) => toast.error(e.message),
  })

  const deleteMut = useMutation({
    mutationFn: (id: string) => deleteUser(id),
    onSuccess: () => { toast.success('User deleted'); setDeleting(null); invalidateAll() },
    onError: (e: Error) => toast.error(e.message),
  })

  const lockMut = useMutation({
    mutationFn: (id: string) => lockUserAccount(id),
    onSuccess: () => { toast.success('Account locked'); setViewing(null); invalidateAll() },
    onError: (e: Error) => toast.error(e.message),
  })

  const unlockMut = useMutation({
    mutationFn: (id: string) => unlockUserAccount(id),
    onSuccess: () => { toast.success('Account unlocked'); setViewing(null); invalidateAll() },
    onError: (e: Error) => toast.error(e.message),
  })

  const toggleStatusMut = useMutation({
    mutationFn: ({ id, currentStatus }: { id: string; currentStatus: string }) =>
      currentStatus === 'active' ? deactivateUser(id) : activateUser(id),
    onSuccess: (_d, vars) => {
      toast.success(vars.currentStatus === 'active' ? 'Account disabled' : 'Account enabled')
      setViewing(null)
      invalidateAll()
    },
    onError: (e: Error) => toast.error(e.message),
  })

  const resetMut = useMutation({
    mutationFn: (id: string) => sendPasswordReset(id),
    onSuccess: () => { toast.success('Password reset email sent'); setViewing(null); invalidateAll() },
    onError: (e: Error) => toast.error(e.message),
  })

  // Columns
  const columns: Column<UserListItem>[] = useMemo(() => {
    const cols: Column<UserListItem>[] = [
      {
        key: 'fullname',
        title: 'User',
        sortable: true,
        render: (u) => (
          <div className="flex items-center gap-3 py-1">
            <div className={clsx(
              'w-9 h-9 rounded-lg flex items-center justify-center text-sm font-semibold shrink-0',
              activeTab === 'system'
                ? 'bg-indigo-500/15 text-indigo-400 border border-indigo-500/20'
                : activeTab === 'org_admin'
                ? 'bg-blue-500/15 text-blue-400 border border-blue-500/20'
                : 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/20'
            )}>
              {(u.fullname?.[0] ?? '?').toUpperCase()}
            </div>
            <div className="min-w-0">
              <p className="text-sm font-medium text-foreground truncate">{u.fullname}</p>
              <p className="text-xs text-foreground/40 truncate">{u.email}</p>
            </div>
          </div>
        ),
      },
      {
        key: 'role',
        title: 'Role',
        sortable: true,
        render: (u) => (
          <span className="text-sm text-foreground/70">{u.role || '—'}</span>
        ),
      },
    ]

    // Add organization column for non-system tabs
    if (activeTab !== 'system') {
      cols.push({
        key: 'organization',
        title: 'Organization',
        render: (u) => (
          <span className="text-sm text-foreground/60">{u.organization || '—'}</span>
        ),
      })
    }

    // Account type badge
    cols.push({
      key: 'account_type' as any,
      title: 'Account Type',
      render: (u) => {
        if (activeTab === 'system') {
          return (
            <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 text-xs font-medium">
              <Shield size={12} /> System
            </span>
          )
        }
        if (activeTab === 'org_admin') {
          return (
            <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-medium">
              <Building2 size={12} /> Org Admin
            </span>
          )
        }
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-medium">
            <UserCheck size={12} /> Employee
          </span>
        )
      },
    })

    // Status column
    cols.push({
      key: 'status',
      title: 'Status',
      render: (u) => (
        <div className="flex items-center gap-1.5">
          <StatusBadge status={toStatus(u.status)} />
          {u.is_locked && (
            <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded-full bg-red-500/10 text-[10px] font-medium text-red-500">
              <Lock size={10} /> Locked
            </span>
          )}
        </div>
      ),
    })

    // Actions column
    cols.push({
      key: 'actions',
      title: '',
      className: 'text-right w-[140px]',
      render: (u) => (
        <div className="flex items-center justify-end gap-1">
          <button
            onClick={(e) => { e.stopPropagation(); setViewing(u) }}
            className="inline-flex h-8 w-8 items-center justify-center rounded-lg text-foreground/40 hover:text-foreground hover:bg-foreground/10 transition-all"
            title="View"
          >
            <Eye size={16} />
          </button>
          {canEditUser && (
            <button
              onClick={(e) => { e.stopPropagation(); setEditing(u) }}
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg text-foreground/40 hover:text-foreground hover:bg-foreground/10 transition-all"
              title="Edit"
            >
              <Pencil size={16} />
            </button>
          )}
          {canDeleteUser && (
            <button
              onClick={(e) => { e.stopPropagation(); setDeleting(u) }}
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg text-foreground/40 hover:text-red-400 hover:bg-red-500/10 transition-all"
              title="Delete"
            >
              <Trash2 size={16} />
            </button>
          )}
        </div>
      ),
    })

    return cols
  }, [activeTab, canEditUser, canDeleteUser])

  const handleSort = (key: string) => {
    setSorts(prev => ({
      ...prev,
      [activeTab]: {
        by: key,
        dir: prev[activeTab].by === key && prev[activeTab].dir === 'asc' ? 'desc' : 'asc',
      },
    }))
    setPage(activeTab, 1)
  }

  const handleCreate = async (v: UserFormValues) => {
    await createMut.mutateAsync({
      fullname: v.fullname,
      email: v.email,
      role_id: v.role_id,
      organization_id: v.organization_id,
      status: v.status,
    })
  }

  const handleEdit = async (v: UserFormValues) => {
    if (!editing) return
    await updateUser(editing.id, {
      fullname: v.fullname,
      email: v.email,
      role_id: v.role_id,
      organization_id: v.organization_id,
      status: v.status,
    })
    toast.success('User updated')
    setEditing(null)
    invalidateAll()
  }

  return (
    <div className="grid gap-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="p-2.5 bg-gradient-to-br from-indigo-500/20 to-purple-500/20 rounded-xl ring-1 ring-indigo-500/20">
            <Users className="w-5 h-5 text-indigo-400" />
          </div>
          <div>
            <h1 className="text-xl font-semibold text-foreground">User Management</h1>
            <p className="text-sm text-foreground/40">
              Manage accounts across all organizations &middot; {(data?.total ?? 0)} users
            </p>
          </div>
        </div>
        {canCreateUser && (
          <button
            onClick={() => setCreating(true)}
            className="btn bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white font-medium flex items-center gap-2 px-4 py-2.5 rounded-xl shadow-lg shadow-blue-500/20 transition-all hover:shadow-blue-500/30"
          >
            <Plus size={16} /> Add User
          </button>
        )}
      </div>

      {/* Main card: tabs + search + table */}
      <div className="rounded-2xl bg-surface-elevated border border-foreground/5 overflow-hidden">
        {/* Tab bar */}
        <div className="flex items-center gap-1 px-4 pt-4 pb-3">
          {TAB_CONFIG.map(tab => {
            const isActive = activeTab === tab.key
            const count = counts[tab.key]
            return (
              <button
                key={tab.key}
                onClick={() => { setActiveTab(tab.key); setPage(tab.key, 1) }}
                className={clsx(
                  'flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium transition-all',
                  isActive
                    ? 'bg-foreground/10 text-foreground shadow-sm'
                    : 'text-foreground/40 hover:text-foreground/70 hover:bg-foreground/5'
                )}
              >
                {tab.icon}
                <span>{tab.label}</span>
                {count > 0 && (
                  <span className={clsx(
                    'px-1.5 py-0.5 rounded-full text-[11px] font-semibold min-w-[20px] text-center',
                    isActive
                      ? 'bg-foreground/10 text-foreground/70'
                      : 'bg-foreground/5 text-foreground/30'
                  )}>
                    {count}
                  </span>
                )}
              </button>
            )
          })}
        </div>

        {/* Search + org filter */}
        <div className="px-4 pb-4">
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="relative flex-1 sm:max-w-sm">
              <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-foreground/30" />
              <input
                type="text"
                placeholder="Search by name or email..."
                value={search}
                onChange={e => { setSearch(e.target.value); setPages({ system: 1, org_admin: 1 }) }}
                className="w-full bg-foreground/5 border border-foreground/10 rounded-xl pl-10 pr-4 py-2.5 text-sm text-foreground placeholder-foreground/30 outline-none focus:border-blue-500/40 focus:ring-1 focus:ring-blue-500/20 transition-all"
              />
            </div>
            {activeTab !== 'system' && (
              <div className="w-full sm:w-64">
                <OrganizationFilter
                  value={orgFilter[activeTab as 'org_admin']}
                  onChange={(v: string) => {
                    setOrgFilter(prev => ({ ...prev, [activeTab]: v }))
                    setPage(activeTab, 1)
                  }}
                />
              </div>
            )}
          </div>
        </div>

        {/* Table */}
        <DataTable<UserListItem>
          columns={columns}
          rows={data?.items ?? []}
          loading={activeQuery.isLoading}
          showIndex
          pagination={{
            page: pages[activeTab],
            pageSize,
            total: data?.total ?? 0,
            onPageChange: (p: number) => setPage(activeTab, p),
            onPageSizeChange: (s: number) => { setPageSize(s); setPage(activeTab, 1) },
          }}
          sort={{
            sortBy: sorts[activeTab].by,
            sortDir: sorts[activeTab].dir,
            onSortChange: handleSort,
          }}
          emptyText="No users found"
        />
      </div>

      {/* Create Modal */}
      <Modal open={creating} onClose={() => setCreating(false)} title="Create User" size="md">
        <UserForm
          mode="create"
          onSubmit={handleCreate}
          onCancel={() => setCreating(false)}
          hideOrganizationSelector={false}
        />
      </Modal>

      {/* Edit Modal */}
      <Modal open={!!editing} onClose={() => setEditing(null)} title="Edit User" size="md">
        {editing && (
          <UserForm
            mode="edit"
            initial={{
              fullname: editing.fullname,
              email: editing.email,
              role_id: editing.role_id,
              organization_id: editing.organization_id,
              status: toStatus(editing.status) as 'active' | 'disabled',
            }}
            onSubmit={handleEdit}
            onCancel={() => setEditing(null)}
            hideOrganizationSelector={false}
          />
        )}
      </Modal>

      {/* View / Actions Modal */}
      <Modal open={!!viewing} onClose={() => setViewing(null)} title="User Details" size="sm">
        {viewing && (
          <div className="space-y-5">
            {/* User card */}
            <div className="flex items-center gap-4 p-4 bg-foreground/5 rounded-xl">
              <div className={clsx(
                'w-12 h-12 rounded-xl flex items-center justify-center text-lg font-bold shrink-0',
                'bg-indigo-500/15 text-indigo-400 border border-indigo-500/20'
              )}>
                {(viewing.fullname?.[0] ?? '?').toUpperCase()}
              </div>
              <div className="min-w-0 flex-1">
                <h3 className="text-lg font-semibold text-foreground truncate">{viewing.fullname}</h3>
                <p className="text-sm text-foreground/40 truncate">{viewing.email}</p>
              </div>
            </div>

            {/* Info grid */}
            <div className="grid grid-cols-2 gap-3">
              <div className="p-3 bg-foreground/5 rounded-xl">
                <p className="text-[10px] uppercase tracking-wider text-foreground/30 mb-1">Role</p>
                <p className="text-sm font-medium text-foreground">{viewing.role || '—'}</p>
              </div>
              <div className="p-3 bg-foreground/5 rounded-xl">
                <p className="text-[10px] uppercase tracking-wider text-foreground/30 mb-1">Organization</p>
                <p className="text-sm font-medium text-foreground">{viewing.organization || '—'}</p>
              </div>
              <div className="p-3 bg-foreground/5 rounded-xl">
                <p className="text-[10px] uppercase tracking-wider text-foreground/30 mb-1">Status</p>
                <StatusBadge status={toStatus(viewing.status)} />
              </div>
              <div className="p-3 bg-foreground/5 rounded-xl">
                <p className="text-[10px] uppercase tracking-wider text-foreground/30 mb-1">Account</p>
                {viewing.is_locked ? (
                  <span className="inline-flex items-center gap-1 text-xs font-medium text-red-500">
                    <Lock size={12} /> Locked
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1 text-xs font-medium text-green-500">
                    <Unlock size={12} /> Unlocked
                  </span>
                )}
              </div>
            </div>

            {/* Actions */}
            <div className="space-y-2 pt-2 border-t border-foreground/5">
              {canEditStatus && (
                toStatus(viewing.status) === 'active' ? (
                  <button
                    onClick={() => toggleStatusMut.mutate({ id: viewing.id, currentStatus: 'active' })}
                    disabled={toggleStatusMut.isPending}
                    className="w-full flex items-center justify-center gap-2 py-2.5 rounded-xl bg-amber-500/10 text-amber-400 font-medium text-sm hover:bg-amber-500/20 transition-colors"
                  >
                    <Power size={15} />
                    {toggleStatusMut.isPending ? 'Disabling...' : 'Disable Account'}
                  </button>
                ) : (
                  <button
                    onClick={() => toggleStatusMut.mutate({ id: viewing.id, currentStatus: 'disabled' })}
                    disabled={toggleStatusMut.isPending}
                    className="w-full flex items-center justify-center gap-2 py-2.5 rounded-xl bg-green-500/10 text-green-400 font-medium text-sm hover:bg-green-500/20 transition-colors"
                  >
                    <Power size={15} />
                    {toggleStatusMut.isPending ? 'Enabling...' : 'Enable Account'}
                  </button>
                )
              )}

              {canUnlock && (
                viewing.is_locked ? (
                  <button
                    onClick={() => unlockMut.mutate(viewing.id)}
                    disabled={unlockMut.isPending}
                    className="w-full flex items-center justify-center gap-2 py-2.5 rounded-xl bg-green-500/10 text-green-400 font-medium text-sm hover:bg-green-500/20 transition-colors"
                  >
                    <Unlock size={15} />
                    {unlockMut.isPending ? 'Unlocking...' : 'Unlock Account'}
                  </button>
                ) : (
                  <button
                    onClick={() => lockMut.mutate(viewing.id)}
                    disabled={lockMut.isPending}
                    className="w-full flex items-center justify-center gap-2 py-2.5 rounded-xl bg-red-500/10 text-red-400 font-medium text-sm hover:bg-red-500/20 transition-colors"
                  >
                    <Lock size={15} />
                    {lockMut.isPending ? 'Locking...' : 'Lock Account'}
                  </button>
                )
              )}

              {canResetPwd && (
                <button
                  onClick={() => resetMut.mutate(viewing.id)}
                  disabled={resetMut.isPending}
                  className="w-full flex items-center justify-center gap-2 py-2.5 rounded-xl bg-blue-500/10 text-blue-400 font-medium text-sm hover:bg-blue-500/20 transition-colors"
                >
                  <Mail size={15} />
                  {resetMut.isPending ? 'Sending...' : 'Send Password Reset'}
                </button>
              )}
            </div>
          </div>
        )}
      </Modal>

      {/* Delete Confirmation */}
      <ConfirmModal
        open={!!deleting}
        onCancel={() => setDeleting(null)}
        onConfirm={async () => { if (deleting) await deleteMut.mutateAsync(deleting.id) }}
        title="Delete User"
        message={deleting ? <>Are you sure you want to delete <b>{deleting.fullname}</b>? This action cannot be undone.</> : ''}
        confirmText={deleteMut.isPending ? 'Deleting...' : 'Delete'}
        danger
      />
    </div>
  )
}
