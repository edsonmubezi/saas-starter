import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import {
  Shield, Search, Save, ChevronDown, ChevronRight,
  Loader2, Layers, CheckSquare, Square, X,
} from 'lucide-react'
import clsx from 'clsx'

import {
  getPermissionsAll,
  updateOrgRolePermissions,
  type PermissionResource,
} from '../../utils/permissions'

import {
  getOrgRoles,
  getRoleDetails,
  type RoleListItem,
  type RoleDetails,
} from '../../utils/roles'

import { useAuth } from '../../state/AuthContext'

const toTitle = (s: string) => s.replace(/_/g, ' ')

export default function OrgAssignRoles() {
  const { user } = useAuth()
  const [selectedRoleId, setSelectedRoleId] = useState<number | string | ''>('')
  const [sel, setSel] = useState<Set<number>>(new Set())
  const [saving, setSaving] = useState(false)
  const [search, setSearch] = useState('')
  const [collapsed, setCollapsed] = useState<Set<number>>(new Set())

  // ── Queries ───────────────────────────────────────────
  const { data: allRoles = [], isFetching: rolesLoading } = useQuery<RoleListItem[]>({
    queryKey: ['roles-all'],
    queryFn: getOrgRoles,
    staleTime: 60_000,
  })

  const roles = useMemo(() => allRoles, [allRoles])

  const { data: allResources = [], isFetching: permsLoading } = useQuery<PermissionResource[]>({
    queryKey: ['permissions-all'],
    queryFn: getPermissionsAll,
    staleTime: 60_000,
  })

  const resources = useMemo(() => {
    return allResources.map(resource => ({
      ...resource,
      permissions: (resource.permissions || []).filter(p => !p.name.startsWith('admin.'))
    })).filter(resource => resource.permissions.length > 0)
  }, [allResources])

  // ── Search / filter ───────────────────────────────────
  const filteredResources = useMemo(() => {
    if (!search.trim()) return resources
    const q = search.toLowerCase()
    return resources
      .map(r => ({
        ...r,
        permissions: r.permissions.filter(p =>
          p.name.toLowerCase().includes(q) ||
          p.description?.toLowerCase().includes(q) ||
          r.resource.toLowerCase().includes(q)
        ),
      }))
      .filter(r => r.permissions.length > 0)
  }, [resources, search])

  // ── Prefill from list data ────────────────────────────
  useEffect(() => {
    if (!selectedRoleId) { setSel(new Set()); return }
    const role = roles.find(r => String(r.id) === String(selectedRoleId))
    if (role?.permission_ids?.length) {
      setSel(new Set(role.permission_ids))
    } else {
      setSel(new Set())
    }
  }, [selectedRoleId, roles])

  // ── Authoritative role details ────────────────────────
  const {
    data: details,
    isFetching: detailsLoading,
    isError: detailsError,
  } = useQuery<RoleDetails>({
    queryKey: ['role-details', selectedRoleId],
    queryFn: () => getRoleDetails(selectedRoleId as string | number),
    enabled: !!selectedRoleId,
    staleTime: 60_000,
  })

  useEffect(() => {
    if (!details) return
    const ids = (details.permissions ?? []).map(p => p.id).filter(n => Number.isFinite(n))
    setSel(new Set(ids))
  }, [details])

  // ── Toggles ───────────────────────────────────────────
  const toggle = useCallback((permId: number, checked: boolean) => {
    setSel(prev => {
      const next = new Set(prev)
      if (checked) next.add(permId); else next.delete(permId)
      return next
    })
  }, [])

  const toggleResource = useCallback((r: PermissionResource, checkAll: boolean) => {
    setSel(prev => {
      const next = new Set(prev)
      for (const p of r.permissions ?? []) {
        if (!p?.id) continue
        if (checkAll) next.add(p.id); else next.delete(p.id)
      }
      return next
    })
  }, [])

  const resourceSelectState = useCallback(
    (r: PermissionResource) => {
      const ids = (r.permissions ?? []).map(p => p.id).filter((x): x is number => Number.isFinite(x))
      const total = ids.length
      const checked = ids.filter(id => sel.has(id)).length
      return { all: total > 0 && checked === total, some: checked > 0 && checked < total, checked, total }
    },
    [sel]
  )

  // ── Collapse ──────────────────────────────────────────
  const toggleCollapse = useCallback((resourceId: number) => {
    setCollapsed(prev => {
      const next = new Set(prev)
      if (next.has(resourceId)) next.delete(resourceId); else next.add(resourceId)
      return next
    })
  }, [])

  const collapseAll = useCallback(() => {
    setCollapsed(new Set(resources.map(r => r.resource_id)))
  }, [resources])

  const expandAll = useCallback(() => {
    setCollapsed(new Set())
  }, [])

  // ── Bulk select ───────────────────────────────────────
  const selectAllVisible = useCallback(() => {
    setSel(prev => {
      const next = new Set(prev)
      for (const r of filteredResources) {
        for (const p of r.permissions) { if (p?.id) next.add(p.id) }
      }
      return next
    })
  }, [filteredResources])

  const deselectAllVisible = useCallback(() => {
    setSel(prev => {
      const next = new Set(prev)
      for (const r of filteredResources) {
        for (const p of r.permissions) { if (p?.id) next.delete(p.id) }
      }
      return next
    })
  }, [filteredResources])

  // ── Stats ─────────────────────────────────────────────
  const totalPermissions = useMemo(() =>
    resources.reduce((acc, r) => acc + r.permissions.length, 0)
  , [resources])

  const selectedCount = sel.size
  const selectedRole = roles.find(r => String(r.id) === String(selectedRoleId))

  // ── Save ──────────────────────────────────────────────
  const handleSave = useCallback(async () => {
    if (!selectedRoleId) { toast.error('Pick a role first'); return }
    try {
      setSaving(true)
      await updateOrgRolePermissions({
        role_id: String(selectedRoleId),
        permission_ids: Array.from(sel),
      })
      toast.success('Role permissions updated successfully')
    } catch (e: any) {
      toast.error(e?.message || e?.response?.data?.error || 'Failed to update role permissions')
    } finally {
      setSaving(false)
    }
  }, [selectedRoleId, sel])

  const isLoading = permsLoading || rolesLoading
  const pctGlobal = totalPermissions > 0 ? Math.round((selectedCount / totalPermissions) * 100) : 0

  // ── Render ────────────────────────────────────────────
  return (
    <div className="grid gap-5">

      {/* ═══ Header ═══ */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="p-2.5 bg-gradient-to-br from-blue-500/20 to-indigo-500/20 rounded-xl ring-1 ring-blue-500/20">
            <Shield className="w-5 h-5 text-blue-400" />
          </div>
          <div>
            <h1 className="text-xl font-semibold text-foreground">Assign Role Permissions</h1>
            <p className="text-sm text-foreground/40">
              {(user as any)?.organization?.name || 'Organization'} &middot; {resources.length} resource groups &middot; {totalPermissions} permissions
            </p>
          </div>
        </div>
      </div>

      {/* ═══ Role Selector Card ═══ */}
      <div className="rounded-2xl bg-surface-elevated border border-foreground/[0.06] overflow-hidden">
        <div className="p-5">
          <div className="flex flex-col lg:flex-row lg:items-center gap-4">
            {/* Role dropdown */}
            <div className="flex-1 min-w-0">
              <label className="block text-[11px] uppercase tracking-wider text-foreground/30 font-medium mb-2">Select Role</label>
              <select
                className="w-full bg-foreground/[0.04] border border-foreground/10 rounded-xl px-4 py-3 text-sm text-foreground
                           outline-none focus:border-blue-500/40 focus:ring-1 focus:ring-blue-500/20 transition-all
                           appearance-none cursor-pointer"
                value={selectedRoleId}
                onChange={e => setSelectedRoleId(e.target.value)}
                disabled={rolesLoading}
              >
                <option value="">{rolesLoading ? 'Loading roles...' : '-- Choose a role to configure --'}</option>
                {roles.map(r => (
                  <option key={String(r.id)} value={String(r.id)}>{r.name}</option>
                ))}
              </select>
            </div>

            {/* Stats */}
            {!!selectedRoleId && (
              <div className="flex items-center gap-3 lg:gap-4">
                {/* Circular progress */}
                <div className="relative flex items-center justify-center w-14 h-14 shrink-0">
                  <svg className="w-14 h-14 -rotate-90" viewBox="0 0 56 56">
                    <circle cx="28" cy="28" r="24" fill="none" stroke="currentColor" strokeWidth="3"
                      className="text-foreground/[0.06]" />
                    <circle cx="28" cy="28" r="24" fill="none" strokeWidth="3"
                      strokeDasharray={`${2 * Math.PI * 24}`}
                      strokeDashoffset={`${2 * Math.PI * 24 * (1 - pctGlobal / 100)}`}
                      strokeLinecap="round"
                      className={clsx(
                        'transition-all duration-700 ease-out',
                        pctGlobal === 100 ? 'text-emerald-400' : pctGlobal > 0 ? 'text-blue-400' : 'text-foreground/20'
                      )}
                      stroke="currentColor"
                    />
                  </svg>
                  <span className="absolute text-xs font-bold text-foreground">{pctGlobal}%</span>
                </div>

                <div className="min-w-0">
                  <p className="text-sm font-semibold text-foreground">
                    {detailsLoading ? (
                      <span className="inline-flex items-center gap-1.5 text-foreground/50">
                        <Loader2 size={13} className="animate-spin" /> Loading...
                      </span>
                    ) : detailsError ? (
                      <span className="text-red-400">Error loading</span>
                    ) : (
                      <>{selectedCount} <span className="text-foreground/40 font-normal">of {totalPermissions}</span></>
                    )}
                  </p>
                  <p className="text-xs text-foreground/30">permissions assigned</p>
                </div>

                {/* Save button */}
                <button
                  className={clsx(
                    'h-10 px-5 rounded-xl text-sm font-medium flex items-center gap-2 transition-all shrink-0',
                    'bg-gradient-to-r from-blue-600 to-blue-500 text-white shadow-lg shadow-blue-500/20',
                    'hover:from-blue-500 hover:to-blue-400 hover:shadow-blue-500/30',
                    'disabled:opacity-50 disabled:cursor-not-allowed disabled:shadow-none'
                  )}
                  onClick={handleSave}
                  disabled={!selectedRoleId || saving || detailsLoading}
                >
                  {saving ? <Loader2 size={15} className="animate-spin" /> : <Save size={15} />}
                  {saving ? 'Saving...' : 'Save'}
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Global progress bar */}
        {!!selectedRoleId && (
          <div className="h-1 bg-foreground/[0.04]">
            <div
              className={clsx(
                'h-full transition-all duration-700 ease-out',
                pctGlobal === 100 ? 'bg-emerald-500' : pctGlobal > 0 ? 'bg-blue-500' : ''
              )}
              style={{ width: `${pctGlobal}%` }}
            />
          </div>
        )}
      </div>

      {/* ═══ Toolbar: Search + Bulk Actions ═══ */}
      {!!selectedRoleId && !detailsLoading && resources.length > 0 && (
        <div className="flex flex-col sm:flex-row sm:items-center gap-3">
          {/* Search */}
          <div className="relative flex-1">
            <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-foreground/25" />
            <input
              type="text"
              placeholder="Filter resources or permissions..."
              className="w-full bg-foreground/[0.04] border border-foreground/[0.08] rounded-xl pl-9 pr-9 py-2.5 text-sm
                         text-foreground placeholder-foreground/25 outline-none
                         focus:border-blue-500/30 focus:ring-1 focus:ring-blue-500/15 transition-all"
              value={search}
              onChange={e => setSearch(e.target.value)}
            />
            {search && (
              <button
                className="absolute right-3 top-1/2 -translate-y-1/2 text-foreground/30 hover:text-foreground/60 transition-colors"
                onClick={() => setSearch('')}
              >
                <X size={14} />
              </button>
            )}
          </div>

          {/* Bulk action buttons */}
          <div className="flex items-center gap-1.5 shrink-0">
            <button
              className="h-9 px-3 text-xs rounded-lg bg-foreground/[0.04] border border-foreground/[0.08] text-foreground/50
                         hover:bg-emerald-500/10 hover:border-emerald-500/20 hover:text-emerald-400 transition-all flex items-center gap-1.5"
              onClick={selectAllVisible}
              title="Select all visible"
            >
              <CheckSquare size={13} /> Select all
            </button>
            <button
              className="h-9 px-3 text-xs rounded-lg bg-foreground/[0.04] border border-foreground/[0.08] text-foreground/50
                         hover:bg-red-500/10 hover:border-red-500/20 hover:text-red-400 transition-all flex items-center gap-1.5"
              onClick={deselectAllVisible}
              title="Deselect all visible"
            >
              <Square size={13} /> Clear all
            </button>
            <div className="w-px h-5 bg-foreground/10 mx-1" />
            <button
              className="h-9 px-3 text-xs rounded-lg bg-foreground/[0.04] border border-foreground/[0.08] text-foreground/50
                         hover:bg-foreground/10 hover:text-foreground/70 transition-all flex items-center gap-1.5"
              onClick={expandAll}
              title="Expand all"
            >
              <ChevronDown size={13} /> Expand
            </button>
            <button
              className="h-9 px-3 text-xs rounded-lg bg-foreground/[0.04] border border-foreground/[0.08] text-foreground/50
                         hover:bg-foreground/10 hover:text-foreground/70 transition-all flex items-center gap-1.5"
              onClick={collapseAll}
              title="Collapse all"
            >
              <ChevronRight size={13} /> Collapse
            </button>
          </div>
        </div>
      )}

      {/* ═══ Search result count ═══ */}
      {search.trim() && (
        <p className="text-xs text-foreground/30 -mt-2">
          Showing {filteredResources.length} of {resources.length} resource groups
          {filteredResources.length === 0 && (
            <span className="ml-1 text-amber-400/60">— try a different search term</span>
          )}
        </p>
      )}

      {/* ═══ Permission Grid ═══ */}
      {!!selectedRoleId && !isLoading && (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
          {filteredResources.map((r) => {
            const { all, some, checked, total } = resourceSelectState(r)
            const isCollapsed = collapsed.has(r.resource_id)
            const pct = total > 0 ? (checked / total) * 100 : 0

            return (
              <div
                key={r.resource_id}
                className={clsx(
                  'rounded-xl border overflow-hidden transition-all duration-300',
                  all
                    ? 'border-emerald-500/25 bg-emerald-500/[0.02]'
                    : some
                    ? 'border-blue-500/20 bg-blue-500/[0.02]'
                    : 'border-foreground/[0.07] bg-foreground/[0.015]'
                )}
              >
                {/* Resource header */}
                <div
                  className="flex items-center gap-2 px-3.5 py-3 cursor-pointer hover:bg-foreground/[0.03] transition-colors select-none"
                  onClick={() => toggleCollapse(r.resource_id)}
                >
                  <span className="text-foreground/30 shrink-0">
                    {isCollapsed ? <ChevronRight size={14} /> : <ChevronDown size={14} />}
                  </span>

                  <span className="text-sm font-semibold capitalize truncate text-foreground flex-1">
                    {toTitle(r.resource)}
                  </span>

                  {/* Count badge */}
                  <span className={clsx(
                    'text-[11px] font-semibold px-2 py-0.5 rounded-md shrink-0 tabular-nums',
                    all
                      ? 'bg-emerald-500/15 text-emerald-400'
                      : some
                      ? 'bg-blue-500/15 text-blue-400'
                      : 'bg-foreground/[0.06] text-foreground/30'
                  )}>
                    {checked}/{total}
                  </span>

                  {/* Select-all checkbox */}
                  <label
                    className="shrink-0 flex items-center"
                    onClick={e => e.stopPropagation()}
                  >
                    <input
                      type="checkbox"
                      className="w-3.5 h-3.5 rounded accent-blue-500 cursor-pointer"
                      checked={all}
                      ref={el => { if (el) el.indeterminate = some && !all }}
                      onChange={e => toggleResource(r, e.currentTarget.checked)}
                      disabled={!selectedRoleId || detailsLoading}
                    />
                  </label>
                </div>

                {/* Mini progress bar */}
                <div className="h-[2px] bg-foreground/[0.04]">
                  <div
                    className={clsx(
                      'h-full transition-all duration-500 ease-out',
                      all ? 'bg-emerald-500/70' : some ? 'bg-blue-500/60' : ''
                    )}
                    style={{ width: `${pct}%` }}
                  />
                </div>

                {/* Permissions list */}
                {!isCollapsed && (
                  <div className="px-2 py-2 grid gap-0.5">
                    {r.permissions.map((p) => {
                      const isChecked = sel.has(p.id)
                      return (
                        <label
                          key={p.id}
                          className={clsx(
                            'flex items-start gap-2.5 rounded-lg px-3 py-2 cursor-pointer transition-all duration-150 group',
                            isChecked
                              ? 'bg-blue-500/[0.08] hover:bg-blue-500/[0.12]'
                              : 'hover:bg-foreground/[0.04]'
                          )}
                        >
                          <input
                            type="checkbox"
                            className={clsx(
                              'mt-[3px] w-3.5 h-3.5 rounded accent-blue-500 cursor-pointer shrink-0 transition-opacity',
                              !isChecked && 'opacity-40 group-hover:opacity-70'
                            )}
                            checked={isChecked}
                            onChange={e => toggle(p.id, e.currentTarget.checked)}
                            disabled={!selectedRoleId || detailsLoading}
                          />
                          <div className="min-w-0">
                            <div className={clsx(
                              'text-[13px] font-medium break-words leading-snug transition-colors',
                              isChecked ? 'text-foreground' : 'text-foreground/60 group-hover:text-foreground/80'
                            )}>
                              {p.name}
                            </div>
                            {p.description && (
                              <div className="text-[11px] text-foreground/30 break-words leading-snug mt-0.5">
                                {p.description}
                              </div>
                            )}
                          </div>
                        </label>
                      )
                    })}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}

      {/* ═══ Loading state ═══ */}
      {isLoading && (
        <div className="rounded-2xl bg-surface-elevated border border-foreground/[0.06] p-12 flex flex-col items-center gap-3">
          <Loader2 className="w-6 h-6 text-blue-400 animate-spin" />
          <p className="text-sm text-foreground/40">Loading permissions...</p>
        </div>
      )}

      {/* ═══ Empty states ═══ */}
      {!isLoading && !selectedRoleId && resources.length > 0 && (
        <div className="rounded-2xl bg-surface-elevated border border-foreground/[0.06] p-12 flex flex-col items-center gap-3 text-center">
          <div className="p-3 rounded-xl bg-foreground/[0.04]">
            <Layers className="w-6 h-6 text-foreground/20" />
          </div>
          <p className="text-sm text-foreground/40">Select a role above to manage its permissions</p>
          <p className="text-xs text-foreground/20">{resources.length} resource groups with {totalPermissions} permissions available</p>
        </div>
      )}

      {!isLoading && resources.length === 0 && (
        <div className="rounded-2xl bg-surface-elevated border border-foreground/[0.06] p-12 flex flex-col items-center gap-3 text-center">
          <div className="p-3 rounded-xl bg-foreground/[0.04]">
            <Shield className="w-6 h-6 text-foreground/20" />
          </div>
          <p className="text-sm text-foreground/40">No permissions available for your organization</p>
        </div>
      )}

      {!rolesLoading && roles.length === 0 && (
        <div className="rounded-2xl bg-surface-elevated border border-foreground/[0.06] p-12 flex flex-col items-center gap-3 text-center">
          <div className="p-3 rounded-xl bg-foreground/[0.04]">
            <Shield className="w-6 h-6 text-foreground/20" />
          </div>
          <p className="text-sm text-foreground/40">No roles available in your organization</p>
        </div>
      )}
    </div>
  )
}
