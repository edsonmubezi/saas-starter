// src/pages/users/OrgAdminUsersPage.tsx
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import {
  Eye, Pencil, Plus, Trash2, Building2, Shield, Users, UserRound,
  Search, ArrowUpRight, Info, Power, Mail, User,
  Lock, Unlock,
} from 'lucide-react'
import toast from 'react-hot-toast'
import clsx from 'clsx'

import {
  getOrgUsers,
  getOrgUserById,
  createOrgUser,
  updateOrgUser,
  deleteOrgUser,
  activateOrgUser,
  deactivateOrgUser,
  lockOrgUserAccount,
  unlockOrgUserAccount,
  sendOrgPasswordReset,
  type UserListItem,
  type UsersResult,
  type CreateUserInput,
} from '../../utils/users'
import { getOrgRoles } from '../../utils/roles'

import Modal from '../../ui/Modal'
import ConfirmModal from '../../ui/ConfirmModal'
import DataTable, { type Column } from '../../ui/DataTable'
import UserForm, { type UserFormValues } from './UserForm'
import StatusBadge from '../../ui/StatusBadge'
import useDebounce from '../../state/useDebounce'
import { useAuth } from '../../state/AuthContext'

type UserTypeTab = '' | 'admin'
type ViewTab = 'details' | 'edit'

/** Determine if user has an admin-level role */
const isAdminRole = (r: UserListItem) => !!r.role && r.role.toLowerCase().includes('admin')

const TABS: { key: UserTypeTab; label: string; icon: React.ReactNode; desc: string }[] = [
  { key: 'admin', label: 'Admins',    icon: <Shield size={16} />, desc: 'Admin accounts' },
  { key: '',      label: 'All Users', icon: <Users size={16} />,  desc: 'All accounts' },
]

const VIEW_TABS: { key: ViewTab; label: string; icon: React.ReactNode }[] = [
  { key: 'details', label: 'Details', icon: <Info size={15} /> },
  { key: 'edit',    label: 'Edit',    icon: <Pencil size={15} /> },
]

export default function OrgAdminUsersPage() {
  const { user } = useAuth()
  const perms: string[] = Array.isArray((user as any)?.permissions) ? (user as any).permissions : []
  const canEditUser = perms.includes('tenant.user.edit')
  const canEditStatus = perms.includes('tenant.user.edit_status')
  const canDeleteUser = perms.includes('tenant.user.delete')
  const canUnlock = perms.includes('tenant.user.unlock')

  // ── State ─────────────────────────────────────────────────
  const [search, setSearch] = useState('')
  const debounced = useDebounce(search, 300)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [userTypeTab, setUserTypeTab] = useState<UserTypeTab>('admin')

  const [openCreate, setOpenCreate] = useState(false)
  const [deleting, setDeleting] = useState<UserListItem | null>(null)

  // View / Edit modal
  const [viewing, setViewing] = useState<UserListItem | null>(null)
  const [viewTab, setViewTab] = useState<ViewTab>('details')
  const [toggling, setToggling] = useState(false)
  const [locking, setLocking] = useState(false)
  const [resetting, setResetting] = useState(false)

  // Promote state
  const [promoting, setPromoting] = useState<UserListItem | null>(null)
  const [promotingRoleId, setPromotingRoleId] = useState<number | undefined>()
  const [promotingLoading, setPromotingLoading] = useState(false)

  // Roles for promote dropdown
  const [roles, setRoles] = useState<{ id: string; name: string }[]>([])
  useEffect(() => {
    getOrgRoles().then(r => setRoles(Array.isArray(r) ? r.map(rr => ({ id: String(rr.id), name: rr.name })) : [])).catch(() => {})
  }, [])

  const hrRoles = roles.filter(r => {
    const n = r.name.toLowerCase()
    return n.includes('admin') || n.includes('manager')
  })

  const qc = useQueryClient()

  // ── List Query ──────────────────────────────────────────────
  const params = useMemo(
    () => ({ q: debounced || undefined, page, page_size: pageSize, user_type: userTypeTab || undefined }),
    [debounced, page, pageSize, userTypeTab],
  )
  const queryKey = useMemo(() => ['orgadmin-users', params] as const, [params])

  const { data, isFetching: loading, error } = useQuery<UsersResult>({
    queryKey,
    queryFn: () => getOrgUsers(params),
    placeholderData: keepPreviousData,
  })

  const items = data?.items ?? []
  const total = data?.total ?? 0
  const pageFromServer = data?.page ?? page
  const pageSizeFromServer = data?.page_size ?? pageSize

  // ── Single user query (for viewing modal) ───────────────────
  const { data: viewUser, isLoading: viewLoading } = useQuery<UserListItem>({
    queryKey: ['org-user', viewing?.id],
    queryFn: () => getOrgUserById(viewing!.id as string),
    enabled: !!viewing,
  })

  // Use fetched user data when available, fall back to list item
  const modalUser = viewUser ?? viewing

  // ── Handlers ──────────────────────────────────────────────
  const refresh = useCallback(
    async (msg?: string) => {
      await qc.invalidateQueries({ queryKey: ['orgadmin-users'] })
      if (viewing) await qc.invalidateQueries({ queryKey: ['org-user', viewing.id] })
      if (msg) toast.success(msg)
    },
    [qc, viewing],
  )

  const handleCreate = async (values: UserFormValues) => {
    const payload: CreateUserInput = {
      fullname: values.fullname.trim(),
      email: values.email.trim(),
      password: values.password ?? '',
      status: values.status,
      role_id: values.role_id,
      organization_id: values.organization_id,
    }
    await createOrgUser(payload)
    setOpenCreate(false)
    setPage(1)
    await refresh('User created successfully')
  }

  const handleDeleteConfirm = async () => {
    if (!deleting) return
    await deleteOrgUser(deleting.id)
    setDeleting(null)
    await refresh('User deleted successfully')
  }

  const handleEditSubmit = async (values: UserFormValues) => {
    if (!viewing) return
    await updateOrgUser(viewing.id, {
      fullname: values.fullname,
      email: values.email,
      status: values.status,
      role_id: values.role_id,
      organization_id: values.organization_id,
    })
    await refresh('User updated successfully')
    setViewTab('details')
  }

  const handleToggleStatus = async () => {
    if (!viewing) return
    try {
      setToggling(true)
      const isActive = modalUser?.status === 'active' || modalUser?.active === true
      if (isActive) await deactivateOrgUser(viewing.id)
      else await activateOrgUser(viewing.id)
      await refresh(isActive ? 'User deactivated' : 'User activated')
    } catch (e: any) {
      toast.error(e?.message || 'Failed to update status')
    } finally {
      setToggling(false)
    }
  }

  const handleToggleLock = async () => {
    if (!viewing) return
    try {
      setLocking(true)
      if (modalUser?.is_locked) {
        await unlockOrgUserAccount(viewing.id)
        await refresh('Account unlocked')
      } else {
        await lockOrgUserAccount(viewing.id)
        await refresh('Account locked')
      }
    } catch (e: any) {
      toast.error(e?.message || 'Failed to update lock status')
    } finally {
      setLocking(false)
    }
  }

  const handlePasswordReset = async () => {
    if (!viewing) return
    try {
      setResetting(true)
      await sendOrgPasswordReset(viewing.id)
      toast.success('Password reset email sent')
    } catch (e: any) {
      toast.error(e?.message || 'Failed to send password reset')
    } finally {
      setResetting(false)
    }
  }

  const handlePromote = async () => {
    if (!promoting || !promotingRoleId) return
    try {
      setPromotingLoading(true)
      await updateOrgUser(promoting.id, {
        fullname: promoting.fullname,
        email: promoting.email,
        status: promoting.status,
        role_id: promotingRoleId,
      })
      setPromoting(null)
      setPromotingRoleId(undefined)
      await refresh('User promoted to HR successfully')
    } catch (e: any) {
      toast.error(e?.message || 'Failed to promote user')
    } finally {
      setPromotingLoading(false)
    }
  }

  const openViewModal = (u: UserListItem) => {
    setViewing(u)
    setViewTab('details')
  }

  const closeViewModal = () => {
    setViewing(null)
    setViewTab('details')
  }

  // ── Columns ───────────────────────────────────────────────
  const columns: Column<UserListItem & { actions?: React.ReactNode }>[] = useMemo(() => [
    {
      key: 'fullname',
      title: 'User',
      sortable: true,
      render: r => (
        <div
          className="flex items-center gap-3 py-1 cursor-pointer group"
          onClick={() => openViewModal(r)}
        >
          <div className={clsx(
            'w-9 h-9 rounded-lg flex items-center justify-center text-sm font-semibold shrink-0',
            isAdminRole(r)
              ? 'bg-blue-500/15 text-blue-400 border border-blue-500/20'
              : 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/20'
          )}>
            {(r.fullname || r.email || '?')[0].toUpperCase()}
          </div>
          <div className="min-w-0">
            <p className="text-sm font-medium text-foreground truncate group-hover:text-blue-400 transition-colors">{r.fullname || '—'}</p>
            <p className="text-xs text-foreground/40 truncate">{r.email || '—'}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'role',
      title: 'Role',
      sortable: true,
      render: r => (
        <span className="text-sm text-foreground/70">{r.role || '—'}</span>
      ),
    },
    {
      key: 'account_type' as any,
      title: 'Account Type',
      render: r => {
        if (isAdminRole(r)) {
          return (
            <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-medium">
              <Shield size={12} /> HR / Admin
            </span>
          )
        }
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-medium">
            <UserRound size={12} /> Employee
          </span>
        )
      },
    },
    {
      key: 'status',
      title: 'Status',
      sortable: true,
      render: r => (
        <div className="flex items-center gap-1.5">
          <StatusBadge status={r.status} />
          {r.is_locked && (
            <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded-full bg-red-500/10 text-[10px] font-medium text-red-500">
              <Lock size={10} /> Locked
            </span>
          )}
        </div>
      ),
    },
    {
      key: 'actions',
      title: 'Actions',
      className: 'text-right',
      render: u => (
        <div className="flex items-center justify-end gap-1.5">
          <button
            className="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-blue-500/15 text-blue-400 hover:bg-blue-500/25 transition-all"
            title="View / Edit"
            onClick={() => openViewModal(u)}
          >
            <Eye size={15} />
          </button>
          {canDeleteUser && (
            <button
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-red-500/15 text-red-400 hover:bg-red-500/25 transition-all"
              title="Delete"
              onClick={() => setDeleting(u)}
            >
              <Trash2 size={15} />
            </button>
          )}
        </div>
      ),
    },
  ], [canDeleteUser])

  // ── Account type badge for modal ─────────────────────────
  const accountTypeBadge = modalUser ? (
    isAdminRole(modalUser) ? (
      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-medium">
        <Shield size={12} /> Admin
      </span>
    ) : (
      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-medium">
        <UserRound size={12} /> User
      </span>
    )
  ) : null

  // ── Render ────────────────────────────────────────────────
  return (
    <div className="grid gap-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="p-2.5 bg-gradient-to-br from-blue-500/20 to-purple-500/20 rounded-xl ring-1 ring-blue-500/20">
            <Building2 className="w-5 h-5 text-blue-400" />
          </div>
          <div>
            <h1 className="text-xl font-semibold text-foreground">Organization Users</h1>
            <p className="text-sm text-foreground/40">
              {(user as any)?.organization?.name || 'Organization'} &middot; {total} users
            </p>
          </div>
        </div>
        <button
          className="btn bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white font-medium flex items-center gap-2 px-4 py-2.5 rounded-xl shadow-lg shadow-blue-500/20
                     transition-all hover:shadow-blue-500/30"
          onClick={() => setOpenCreate(true)}
        >
          <Plus size={16} /> Add User
        </button>
      </div>

      {/* Tabs + Search */}
      <div className="rounded-2xl bg-surface-elevated border border-foreground/5 overflow-hidden">
        {/* Tab bar */}
        <div className="flex items-center gap-1 px-4 pt-4 pb-3">
          {TABS.map(tab => (
            <button
              key={tab.key}
              onClick={() => { setUserTypeTab(tab.key); setPage(1) }}
              className={clsx(
                'flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium transition-all',
                userTypeTab === tab.key
                  ? 'bg-foreground/10 text-foreground shadow-sm'
                  : 'text-foreground/40 hover:text-foreground/70 hover:bg-foreground/5'
              )}
            >
              {tab.icon}
              <span>{tab.label}</span>
            </button>
          ))}
        </div>

        {/* Search bar */}
        <div className="px-4 pb-4">
          <div className="relative">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-foreground/30" />
            <input
              className="w-full sm:max-w-sm bg-foreground/5 border border-foreground/10 rounded-xl pl-10 pr-4 py-2.5 text-sm
                         text-foreground placeholder-foreground/30 outline-none focus:border-blue-500/40 focus:ring-1 focus:ring-blue-500/20 transition-all"
              placeholder="Search by name or email..."
              value={search}
              onChange={e => { setSearch(e.target.value); setPage(1) }}
            />
          </div>
        </div>

        {/* Error */}
        {error && (
          <div className="mx-4 mb-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 p-4 text-sm">
            {(error as any)?.message || 'Failed to load users'}
          </div>
        )}

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
            onPageSizeChange: s => { setPageSize(s); setPage(1) },
          }}
        />
      </div>

      {/* ─── Create modal ─── */}
      <Modal open={openCreate} onClose={() => setOpenCreate(false)} title="Add User to Organization">
        <UserForm
          mode="create"
          hideOrganizationSelector
          onSubmit={async v => { try { await handleCreate(v) } catch (e: any) { toast.error(e?.message || 'Failed to create user') } }}
          onCancel={() => setOpenCreate(false)}
        />
      </Modal>

      {/* ─── View / Edit modal ─── */}
      <Modal open={!!viewing} onClose={closeViewModal} title={modalUser?.fullname || 'User Details'} size="lg">
        {viewing && (
          <div className="grid gap-5">
            {/* Modal header with avatar + status */}
            <div className="flex items-center gap-3">
              <div className={clsx(
                'h-12 w-12 rounded-xl flex items-center justify-center text-lg font-bold shrink-0',
                modalUser && isAdminRole(modalUser)
                  ? 'bg-blue-500/15 text-blue-400 border border-blue-500/20'
                  : 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/20'
              )}>
                {(modalUser?.fullname || modalUser?.email || '?')[0].toUpperCase()}
              </div>
              <div className="min-w-0 flex-1">
                <h2 className="text-base font-semibold text-foreground truncate">{modalUser?.fullname || '—'}</h2>
                <p className="text-xs text-foreground/40 truncate">{modalUser?.email}</p>
              </div>
              {modalUser && <StatusBadge status={modalUser.status} size="md" />}
            </div>

            {/* Tab switcher */}
            <div className="flex gap-1 p-1 rounded-xl bg-foreground/[0.04] border border-foreground/[0.06] w-fit">
              {VIEW_TABS.filter(t => t.key !== 'edit' || canEditUser).map(t => (
                <button
                  key={t.key}
                  onClick={() => setViewTab(t.key)}
                  className={clsx(
                    'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all',
                    viewTab === t.key
                      ? 'bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-lg shadow-blue-500/20'
                      : 'text-foreground/40 hover:text-foreground/70 hover:bg-foreground/5'
                  )}
                >
                  {t.icon}
                  <span>{t.label}</span>
                </button>
              ))}
            </div>

            {/* Loading spinner */}
            {viewLoading && (
              <div className="flex items-center justify-center py-8">
                <div className="animate-spin h-6 w-6 border-2 border-blue-500 border-t-transparent rounded-full" />
              </div>
            )}

            {/* Details tab */}
            {viewTab === 'details' && modalUser && !viewLoading && (
              <div className="grid gap-5">
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                  <div className="p-4 rounded-xl bg-foreground/[0.03] border border-foreground/5">
                    <div className="flex items-center gap-2 mb-2">
                      <User size={14} className="text-foreground/30" />
                      <p className="text-[11px] text-foreground/30 uppercase tracking-wider">Full Name</p>
                    </div>
                    <p className="text-sm font-medium text-foreground">{modalUser.fullname || '—'}</p>
                  </div>

                  <div className="p-4 rounded-xl bg-foreground/[0.03] border border-foreground/5">
                    <div className="flex items-center gap-2 mb-2">
                      <Mail size={14} className="text-foreground/30" />
                      <p className="text-[11px] text-foreground/30 uppercase tracking-wider">Email</p>
                    </div>
                    <p className="text-sm font-medium text-foreground truncate">{modalUser.email}</p>
                  </div>

                  <div className="p-4 rounded-xl bg-foreground/[0.03] border border-foreground/5">
                    <div className="flex items-center gap-2 mb-2">
                      <Shield size={14} className="text-foreground/30" />
                      <p className="text-[11px] text-foreground/30 uppercase tracking-wider">Role</p>
                    </div>
                    <p className="text-sm font-medium text-foreground">{modalUser.role || '—'}</p>
                  </div>

                  <div className="p-4 rounded-xl bg-foreground/[0.03] border border-foreground/5">
                    <p className="text-[11px] text-foreground/30 uppercase tracking-wider mb-2">Account Type</p>
                    {accountTypeBadge}
                  </div>

                  <div className="p-4 rounded-xl bg-foreground/[0.03] border border-foreground/5">
                    <p className="text-[11px] text-foreground/30 uppercase tracking-wider mb-2">Status</p>
                    <StatusBadge status={modalUser.status} size="md" />
                  </div>

                  <div className="p-4 rounded-xl bg-foreground/[0.03] border border-foreground/5">
                    <p className="text-[11px] text-foreground/30 uppercase tracking-wider mb-2">Account</p>
                    {modalUser.is_locked ? (
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
                <div className="flex flex-wrap items-center gap-3 pt-4 border-t border-foreground/5">
                  {canEditUser && (
                    <button
                      className="h-9 px-4 text-xs rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400
                                 hover:bg-blue-500/20 transition-all disabled:opacity-50 flex items-center gap-2"
                      disabled={resetting}
                      onClick={handlePasswordReset}
                    >
                      <Mail size={14} />
                      {resetting ? 'Sending...' : 'Send Password Reset'}
                    </button>
                  )}

                  {modalUser && !isAdminRole(modalUser) && canEditUser && (
                    <button
                      className="h-9 px-4 text-xs rounded-lg bg-purple-500/10 border border-purple-500/20 text-purple-400
                                 hover:bg-purple-500/20 transition-all flex items-center gap-2"
                      onClick={() => { setPromoting(viewing); setPromotingRoleId(undefined) }}
                    >
                      <ArrowUpRight size={14} /> Promote to HR
                    </button>
                  )}

                  {canEditStatus && (
                    <button
                      className={clsx(
                        'h-9 px-4 text-xs rounded-lg border transition-all disabled:opacity-50 flex items-center gap-2',
                        modalUser.status === 'active' || modalUser.active === true
                          ? 'bg-red-500/10 border-red-500/20 text-red-400 hover:bg-red-500/20'
                          : 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400 hover:bg-emerald-500/20'
                      )}
                      disabled={toggling}
                      onClick={handleToggleStatus}
                    >
                      <Power size={14} />
                      {toggling ? 'Working...' : (modalUser.status === 'active' || modalUser.active === true) ? 'Deactivate' : 'Activate'}
                    </button>
                  )}

                  {canUnlock && (
                    <button
                      className={clsx(
                        'h-9 px-4 text-xs rounded-lg border transition-all disabled:opacity-50 flex items-center gap-2',
                        modalUser.is_locked
                          ? 'bg-green-500/10 border-green-500/20 text-green-400 hover:bg-green-500/20'
                          : 'bg-red-500/10 border-red-500/20 text-red-400 hover:bg-red-500/20'
                      )}
                      disabled={locking}
                      onClick={handleToggleLock}
                    >
                      {modalUser.is_locked ? <Unlock size={14} /> : <Lock size={14} />}
                      {locking ? 'Working...' : modalUser.is_locked ? 'Unlock' : 'Lock'}
                    </button>
                  )}

                  {canEditUser && (
                    <button
                      className="h-9 px-4 text-xs rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400
                                 hover:bg-blue-500/20 transition-all flex items-center gap-2"
                      onClick={() => setViewTab('edit')}
                    >
                      <Pencil size={14} /> Edit User
                    </button>
                  )}
                </div>
              </div>
            )}

            {/* Edit tab */}
            {viewTab === 'edit' && canEditUser && modalUser && !viewLoading && (
              <div className="max-w-2xl">
                <UserForm
                  mode="edit"
                  initial={modalUser}
                  hideOrganizationSelector
                  onSubmit={async v => {
                    try { await handleEditSubmit(v) }
                    catch (e: any) { toast.error(e?.message || 'Failed to update user') }
                  }}
                  onCancel={() => setViewTab('details')}
                />
              </div>
            )}
          </div>
        )}
      </Modal>

      {/* ─── Promote to HR modal ─── */}
      <Modal open={!!promoting} onClose={() => setPromoting(null)} title="Promote to HR" size="md">
        {promoting && (
          <div className="grid gap-5">
            <p className="text-sm text-foreground/60">
              Promote <span className="text-foreground font-medium">{promoting.fullname}</span> ({promoting.email})
              to an admin role. They will gain access to the organization admin portal.
            </p>

            <label className="grid gap-2">
              <span className="text-sm text-foreground/70">Select HR Role</span>
              <select
                className="select bg-foreground/5 border border-foreground/10 rounded-xl px-4 py-2.5 text-sm text-foreground"
                value={promotingRoleId ?? ''}
                onChange={e => setPromotingRoleId(e.target.value ? Number(e.target.value) : undefined)}
              >
                <option value="">— Choose a role —</option>
                {(hrRoles.length > 0 ? hrRoles : roles).map(r => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </select>
            </label>

            <div className="flex justify-end gap-2 pt-2 border-t border-foreground/5">
              <button
                className="h-9 px-4 text-sm rounded-lg bg-foreground/5 border border-foreground/10 text-foreground/60 hover:bg-foreground/10 transition-all"
                onClick={() => setPromoting(null)}
              >
                Cancel
              </button>
              <button
                className="h-9 px-5 text-sm rounded-lg bg-gradient-to-r from-purple-600 to-blue-600 text-white font-medium
                           hover:from-purple-500 hover:to-blue-500 transition-all disabled:opacity-50 flex items-center gap-2"
                disabled={!promotingRoleId || promotingLoading}
                onClick={handlePromote}
              >
                {promotingLoading ? 'Promoting...' : <><ArrowUpRight size={15} /> Promote</>}
              </button>
            </div>
          </div>
        )}
      </Modal>

      {/* ─── Delete modal ─── */}
      <ConfirmModal
        open={!!deleting}
        onCancel={() => setDeleting(null)}
        onConfirm={async () => { try { await handleDeleteConfirm() } catch (e: any) { toast.error(e?.message || 'Failed to delete user') } }}
        title="Delete User"
        message={deleting ? <>Are you sure you want to delete <b>{deleting.email}</b>?</> : ''}
        confirmText="Delete"
        danger
      />
    </div>
  )
}
