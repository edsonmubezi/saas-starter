import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import toast from 'react-hot-toast'

import {
  AdmingetPermissionsAll,
  updateRolePermissions,
  type PermissionResource,
} from '../../utils/permissions'

import {
  getRoles,
  getRoleDetails,
  type RoleListItem,
  type RoleDetails,
} from '../../utils/roles'

const toTitle = (s: string) => s.replace(/_/g, ' ')

export default function AssignRolePermissionsPage() {
  const [selectedRoleId, setSelectedRoleId] = useState<number | string | ''>('')
  const [sel, setSel] = useState<Set<number>>(new Set())
  const [saving, setSaving] = useState(false)

  // All roles (non-paged)
  const { data: roles = [], isFetching: rolesLoading } = useQuery<RoleListItem[]>({
    queryKey: ['roles-all'],
    queryFn: getRoles,
    staleTime: 60_000,
  })

  // All permissions (non-paged)
  const { data: resources = [], isFetching: permsLoading } = useQuery<PermissionResource[]>({
    queryKey: ['permissions-all'],
    queryFn: AdmingetPermissionsAll,
    staleTime: 60_000,
  })

  // Fast prefill (optional): use the list's permission_ids if present, so the UI feels instant
  useEffect(() => {
    if (!selectedRoleId) {
      setSel(new Set())
      return
    }
    const role = roles.find(r => String(r.id) === String(selectedRoleId))
    if (role?.permission_ids?.length) {
      setSel(new Set(role.permission_ids))
    } else {
      setSel(new Set()) // will be set properly by details below
    }
  }, [selectedRoleId, roles])

  // Role details (authoritative) — fetch when a role is selected
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

  // When details arrive, normalize sel from details.permissions (authoritative)
  useEffect(() => {
    if (!details) return
    const ids = (details.permissions ?? []).map(p => p.id).filter(n => Number.isFinite(n))
    setSel(new Set(ids))
  }, [details])

  // Toggle one permission
  const toggle = useCallback((permId: number, checked: boolean) => {
    setSel(prev => {
      const next = new Set(prev)
      if (checked) next.add(permId)
      else next.delete(permId)
      return next
    })
  }, [])

  // Toggle all within a resource
  const toggleResource = useCallback((r: PermissionResource, checkAll: boolean) => {
    setSel(prev => {
      const next = new Set(prev)
      for (const p of r.permissions ?? []) {
        if (!p?.id) continue
        if (checkAll) next.add(p.id)
        else next.delete(p.id)
      }
      return next
    })
  }, [])

  // Resource checkbox state (all / some)
  const resourceSelectState = useCallback(
    (r: PermissionResource) => {
      const ids = (r.permissions ?? []).map(p => p.id).filter((x): x is number => Number.isFinite(x))
      const total = ids.length
      const checked = ids.filter(id => sel.has(id)).length
      return { all: total > 0 && checked === total, some: checked > 0 && checked < total }
    },
    [sel]
  )

  const handleSave = useCallback(async () => {
    if (!selectedRoleId) {
      toast.error('Pick a role first')
      return
    }
    // It’s valid to save zero permissions too (this implies removing all).
    try {
      setSaving(true)
      await updateRolePermissions({
        role_id: String(selectedRoleId),
        permission_ids: Array.from(sel),
      })
      toast.success('Permissions updated')
    } catch (e: any) {
      const msg = e?.message || e?.response?.data?.error || 'Failed to update permissions'
      toast.error(msg)
    } finally {
      setSaving(false)
    }
  }, [selectedRoleId, sel])

  const totalResources = resources.length

  return (
    <div className="grid gap-5">
      {/* Header */}
      <div className="rounded-2xl border border-foreground/10 bg-gradient-to-br from-indigo-600/30 via-violet-600/20 to-fuchsia-600/10 p-6">
        <h1 className="text-2xl md:text-3xl font-semibold">Assign permissions to role</h1>
        <p className="text-foreground/70 mt-1">
          Pick a role, then tick the permissions to grant. Lists are grouped by resource (3 columns).
        </p>
      </div>

      {/* Toolbar: Role dropdown + Save */}
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-foreground/10 bg-foreground/5 px-4 py-3">
        <div className="flex items-center gap-3">
          <label className="text-sm text-foreground/70">Role</label>
          <select
            className="input h-9 w-full sm:w-auto sm:min-w-[220px]"
            value={selectedRoleId}
            onChange={e => setSelectedRoleId(e.target.value)}
            disabled={rolesLoading}
          >
            <option value="">{rolesLoading ? 'Loading roles…' : 'Select role…'}</option>
            {roles.map(r => (
              <option key={String(r.id)} value={String(r.id)}>
                {r.name}
              </option>
            ))}
          </select>
          {!!selectedRoleId && (
            <span className="text-xs text-foreground/60">
              {detailsLoading ? 'Loading role permissions…' : detailsError ? 'Failed to load role permissions' : 'Permissions loaded'}
            </span>
          )}
        </div>

        <div className="flex items-center gap-2">
          <span className="text-sm text-foreground/60">
            Total resources: <span className="text-foreground">{totalResources}</span>
          </span>
          <button
            className="inline-flex h-9 items-center justify-center rounded-lg px-3 text-sm font-medium
                       bg-emerald-600 hover:bg-emerald-700 text-white border border-transparent disabled:opacity-60"
            onClick={handleSave}
            disabled={!selectedRoleId || saving || permsLoading || detailsLoading}
            title="Save selected permissions to this role"
          >
            {saving ? 'Saving…' : 'Save changes'}
          </button>
        </div>
      </div>

      {/* Permissions grid (3 columns, grouped by resource) */}
      <div className="rounded-xl border border-foreground/10">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 divide-y md:divide-y-0 md:divide-x divide-white/10">
          {(resources ?? []).map((r) => {
            const { all, some } = resourceSelectState(r)
            return (
              <div key={r.resource_id} className="p-4">
                {/* Resource header + Select all */}
                <div className="mb-3 flex items-center justify-between gap-2">
                  <div className="font-semibold capitalize truncate">{toTitle(r.resource)}</div>
                  <label className="inline-flex items-center gap-2 text-xs">
                    <input
                      type="checkbox"
                      className="accent-indigo-500"
                      checked={all}
                      ref={el => { if (el) el.indeterminate = some && !all }}
                      onChange={e => toggleResource(r, e.currentTarget.checked)}
                      disabled={!selectedRoleId || detailsLoading}
                    />
                    <span className="text-foreground/70">Select all</span>
                  </label>
                </div>

                {/* Permission list */}
                <div className="grid gap-2">
                  {(r.permissions ?? []).map((p) => {
                    const checked = sel.has(p.id)
                    return (
                      <label
                        key={p.id}
                        className="flex items-start gap-2 rounded-lg border border-foreground/10 bg-foreground/5 px-3 py-2"
                      >
                        <input
                          type="checkbox"
                          className="mt-0.5 accent-indigo-500"
                          checked={checked}
                          onChange={e => toggle(p.id, e.currentTarget.checked)}
                          disabled={!selectedRoleId || detailsLoading}
                        />
                        <div className="min-w-0">
                          <div className="text-sm font-medium break-words">{p.name}</div>
                          {p.description && (
                            <div className="text-[12px] text-foreground/60 break-words">{p.description}</div>
                          )}
                        </div>
                      </label>
                    )
                  })}
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
