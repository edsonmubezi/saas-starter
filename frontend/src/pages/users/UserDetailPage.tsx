// src/pages/users/UserDetailPage.tsx
import React, { useCallback, useEffect, useState } from 'react'
import { useParams, useNavigate, useSearchParams } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  ArrowLeft, Shield, UserRound, ArrowUpRight, KeyRound, Power,
  Pencil, Info, Mail, User, History,
} from 'lucide-react'
import toast from 'react-hot-toast'
import clsx from 'clsx'

import {
  getOrgUserById,
  updateOrgUser,
  activateOrgUser,
  deactivateOrgUser,
  type UserListItem,
} from '../../utils/users'
import { getOrgRoles } from '../../utils/roles'
import UserForm, { type UserFormValues } from './UserForm'
import StatusBadge from '../../ui/StatusBadge'
import Modal from '../../ui/Modal'
import { useAuth } from '../../state/AuthContext'
import AdvancedPasswordResetModal from '../../components/AdvancedPasswordResetModal'
import PasswordResetHistoryModal from '../../components/PasswordResetHistoryModal'

const isAdminRole = (r: UserListItem) => !!r.role && r.role.toLowerCase().includes('admin')

type Tab = 'details' | 'edit'

const TAB_CONFIG: { key: Tab; label: string; icon: React.ReactNode }[] = [
  { key: 'details', label: 'User Details', icon: <Info size={15} /> },
  { key: 'edit',    label: 'Edit',         icon: <Pencil size={15} /> },
]

export default function UserDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const qc = useQueryClient()
  const { user: authUser } = useAuth()
  const perms: string[] = Array.isArray((authUser as any)?.permissions) ? (authUser as any).permissions : []
  const canEditUser   = perms.includes('tenant.user.edit')
  const canEditStatus = perms.includes('tenant.user.edit_status')

  const initialTab = (searchParams.get('tab') as Tab) || 'details'
  const [tab, setTab] = useState<Tab>(initialTab)
  const [toggling, setToggling] = useState(false)
  const [showResetModal, setShowResetModal] = useState(false)
  const [showHistoryModal, setShowHistoryModal] = useState(false)

  // Promote state
  const [showPromote, setShowPromote] = useState(false)
  const [promotingRoleId, setPromotingRoleId] = useState<number | undefined>()
  const [promotingLoading, setPromotingLoading] = useState(false)
  const [roles, setRoles] = useState<{ id: string; name: string }[]>([])

  useEffect(() => {
    getOrgRoles()
      .then(r => setRoles(Array.isArray(r) ? r.map(rr => ({ id: String(rr.id), name: rr.name })) : []))
      .catch(() => {})
  }, [])

  const hrRoles = roles.filter(r => {
    const n = r.name.toLowerCase()
    return n.includes('admin') || n.includes('hr') || n.includes('manager')
  })

  const { data: user, isLoading, isError } = useQuery<UserListItem>({
    queryKey: ['org-user', id],
    queryFn: () => getOrgUserById(id!),
    enabled: !!id,
  })

  const refresh = useCallback(async () => {
    await qc.invalidateQueries({ queryKey: ['org-user', id] })
    await qc.invalidateQueries({ queryKey: ['orgadmin-users'] })
  }, [qc, id])

  // ── Handlers ────────────────────────────────────────
  const handleEditSubmit = async (values: UserFormValues) => {
    if (!user) return
    await updateOrgUser(user.id, {
      fullname: values.fullname,
      email: values.email,
      status: values.status,
      role_id: values.role_id,
      organization_id: values.organization_id,
    })
    await refresh()
    toast.success('User updated successfully')
    setTab('details')
  }

  const handleToggleStatus = async () => {
    if (!user) return
    try {
      setToggling(true)
      const isActive = user.status === 'active' || user.active === true
      if (isActive) await deactivateOrgUser(user.id)
      else await activateOrgUser(user.id)
      await refresh()
      toast.success(isActive ? 'User deactivated' : 'User activated')
    } catch (e: any) {
      toast.error(e?.message || 'Failed to update status')
    } finally {
      setToggling(false)
    }
  }

  const handlePromote = async () => {
    if (!user || !promotingRoleId) return
    try {
      setPromotingLoading(true)
      await updateOrgUser(user.id, {
        fullname: user.fullname,
        email: user.email,
        status: user.status,
        role_id: promotingRoleId,
      })
      setShowPromote(false)
      setPromotingRoleId(undefined)
      await refresh()
      toast.success('User promoted to HR successfully')
    } catch (e: any) {
      toast.error(e?.message || 'Failed to promote user')
    } finally {
      setPromotingLoading(false)
    }
  }

  // ── Loading / Error ─────────────────────────────────
  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="animate-spin h-8 w-8 border-2 border-blue-500 border-t-transparent rounded-full" />
      </div>
    )
  }

  if (isError || !user) {
    return (
      <div className="rounded-xl border border-foreground/10 bg-surface-elevated p-8 text-center">
        <p className="text-foreground/50 mb-4">Failed to load user.</p>
        <button
          className="px-4 py-2 rounded-lg bg-foreground/10 text-sm hover:bg-foreground/15 transition-colors"
          onClick={() => navigate('/org-users')}
        >
          Back to users
        </button>
      </div>
    )
  }

  const isActive = user.status === 'active' || user.active === true
  const isAdmin = isAdminRole(user)

  const accountTypeBadge = isAdmin ? (
    <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-medium">
      <Shield size={12} /> HR / Admin
    </span>
  ) : (
    <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-medium">
      <UserRound size={12} /> Employee
    </span>
  )

  return (
    <div className="grid gap-5">
      {/* ── Header ── */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/org-users')}
          className="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-foreground/[0.06] border border-foreground/10 hover:bg-foreground/10 hover:border-foreground/20 transition-all"
          title="Back to users"
        >
          <ArrowLeft size={18} className="text-foreground/70" />
        </button>
        <div className="flex items-center gap-3 min-w-0 flex-1">
          <div className={clsx(
            'h-12 w-12 rounded-xl flex items-center justify-center text-lg font-bold shrink-0',
            isAdmin
              ? 'bg-blue-500/15 text-blue-400 border border-blue-500/20'
              : 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/20'
          )}>
            {(user.fullname || user.email || '?')[0].toUpperCase()}
          </div>
          <div className="min-w-0">
            <h1 className="text-lg font-semibold text-foreground truncate">{user.fullname || '—'}</h1>
            <p className="text-xs text-foreground/40 truncate">{user.email}</p>
          </div>
        </div>
        <StatusBadge status={user.status} size="md" />
      </div>

      {/* ── Tabs ── */}
      <div className="flex gap-1 p-1 rounded-xl bg-foreground/[0.04] border border-foreground/[0.06] w-fit">
        {TAB_CONFIG.filter(t => t.key !== 'edit' || canEditUser).map(t => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={clsx(
              'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all',
              tab === t.key
                ? 'bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-lg shadow-blue-500/20'
                : 'text-foreground/40 hover:text-foreground/70 hover:bg-foreground/5'
            )}
          >
            {t.icon}
            <span>{t.label}</span>
          </button>
        ))}
      </div>

      {/* ── Tab Content ── */}
      <div className="rounded-2xl bg-surface-elevated border border-foreground/5 overflow-hidden">
        {tab === 'details' && (
          <div className="p-6 grid gap-6">
            {/* Info grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <div className="p-4 rounded-xl bg-foreground/[0.03] border border-foreground/5">
                <div className="flex items-center gap-2 mb-2">
                  <User size={14} className="text-foreground/30" />
                  <p className="text-[11px] text-foreground/30 uppercase tracking-wider">Full Name</p>
                </div>
                <p className="text-sm font-medium text-foreground">{user.fullname || '—'}</p>
              </div>

              <div className="p-4 rounded-xl bg-foreground/[0.03] border border-foreground/5">
                <div className="flex items-center gap-2 mb-2">
                  <Mail size={14} className="text-foreground/30" />
                  <p className="text-[11px] text-foreground/30 uppercase tracking-wider">Email</p>
                </div>
                <p className="text-sm font-medium text-foreground truncate">{user.email}</p>
              </div>

              <div className="p-4 rounded-xl bg-foreground/[0.03] border border-foreground/5">
                <div className="flex items-center gap-2 mb-2">
                  <Shield size={14} className="text-foreground/30" />
                  <p className="text-[11px] text-foreground/30 uppercase tracking-wider">Role</p>
                </div>
                <p className="text-sm font-medium text-foreground">{user.role || '—'}</p>
              </div>

              <div className="p-4 rounded-xl bg-foreground/[0.03] border border-foreground/5">
                <p className="text-[11px] text-foreground/30 uppercase tracking-wider mb-2">Account Type</p>
                {accountTypeBadge}
              </div>

              <div className="p-4 rounded-xl bg-foreground/[0.03] border border-foreground/5">
                <p className="text-[11px] text-foreground/30 uppercase tracking-wider mb-2">Status</p>
                <StatusBadge status={user.status} size="md" />
              </div>

            </div>

            {/* Actions */}
            <div className="flex flex-wrap items-center gap-3 pt-4 border-t border-foreground/5">
              {canEditUser && (
                <button
                  className="h-9 px-4 text-xs rounded-lg bg-orange-500/10 border border-orange-500/20 text-orange-400
                             hover:bg-orange-500/20 transition-all flex items-center gap-2"
                  onClick={() => setShowResetModal(true)}
                >
                  <KeyRound size={14} />
                  Reset Password
                </button>
              )}

              {canEditUser && (
                <button
                  className="h-9 px-4 text-xs rounded-lg bg-foreground/[0.06] border border-foreground/10 text-foreground/60
                             hover:bg-foreground/10 transition-all flex items-center gap-2"
                  onClick={() => setShowHistoryModal(true)}
                >
                  <History size={14} />
                  Reset History
                </button>
              )}

              {!isAdmin && canEditUser && (
                <button
                  className="h-9 px-4 text-xs rounded-lg bg-purple-500/10 border border-purple-500/20 text-purple-400
                             hover:bg-purple-500/20 transition-all flex items-center gap-2"
                  onClick={() => { setShowPromote(true); setPromotingRoleId(undefined) }}
                >
                  <ArrowUpRight size={14} /> Promote to HR
                </button>
              )}

              {canEditStatus && (
                <button
                  className={clsx(
                    'h-9 px-4 text-xs rounded-lg border transition-all disabled:opacity-50 flex items-center gap-2',
                    isActive
                      ? 'bg-red-500/10 border-red-500/20 text-red-400 hover:bg-red-500/20'
                      : 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400 hover:bg-emerald-500/20'
                  )}
                  disabled={toggling}
                  onClick={handleToggleStatus}
                >
                  <Power size={14} />
                  {toggling ? 'Working...' : isActive ? 'Deactivate' : 'Activate'}
                </button>
              )}

              {canEditUser && (
                <button
                  className="h-9 px-4 text-xs rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400
                             hover:bg-blue-500/20 transition-all flex items-center gap-2"
                  onClick={() => setTab('edit')}
                >
                  <Pencil size={14} /> Edit User
                </button>
              )}
            </div>
          </div>
        )}

        {tab === 'edit' && canEditUser && (
          <div className="p-6 max-w-2xl">
            <UserForm
              mode="edit"
              initial={user}
              hideOrganizationSelector
              onSubmit={async v => {
                try { await handleEditSubmit(v) }
                catch (e: any) { toast.error(e?.message || 'Failed to update user') }
              }}
              onCancel={() => setTab('details')}
            />
          </div>
        )}
      </div>

      {/* ── Password Reset Modals ── */}
      {user && (
        <>
          <AdvancedPasswordResetModal
            userId={user.id}
            userName={user.fullname || user.email}
            isOpen={showResetModal}
            onClose={() => setShowResetModal(false)}
            orgLevel
          />
          <PasswordResetHistoryModal
            userId={user.id}
            userName={user.fullname || user.email}
            isOpen={showHistoryModal}
            onClose={() => setShowHistoryModal(false)}
            orgLevel
          />
        </>
      )}

      {/* ── Promote to HR modal (keep as modal since it's a quick action) ── */}
      <Modal open={showPromote} onClose={() => setShowPromote(false)} title="Assign Admin Role" size="md">
        <div className="grid gap-5">
          <p className="text-sm text-foreground/60">
            Promote <span className="text-foreground font-medium">{user.fullname}</span> ({user.email})
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
              onClick={() => setShowPromote(false)}
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
      </Modal>
    </div>
  )
}
