import { request, unwrap } from './common'
import { API_PREFIX } from './apiPrefix'

// Route prefixes for different user contexts
const ADMIN_PERMISSIONS = `${API_PREFIX.ADMIN}/permissions`
const ORG_PERMISSIONS = `${API_PREFIX.ORG}/permissions`
const ADMIN_ROLES = `${API_PREFIX.ADMIN}/roles`
const ORG_ROLES = `${API_PREFIX.ORG}/roles`

export type PermissionItem = {
  id: number
  name: string
  description?: string
  visibility?: 'ALL' | 'HQ_ONLY' | string
}

export type PermissionResource = {
  resource_id: number
  resource: string
  permissions: PermissionItem[]
}

export type GetPermissionResourcesParams = {
  page?: number
  page_size?: number
}

export type PermissionResourcesResult = {
  items: PermissionResource[]
  total: number
  page: number
  page_size: number
  total_pages?: number
  has_next?: boolean
  has_prev?: boolean
}

const qsFrom = (params: Record<string, unknown>) => {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return
    qs.set(k, String(v))
  })
  return qs.toString()
}

const mapPermission = (p: any): PermissionItem => ({
  id: Number(p.id),
  name: String(p.name ?? ''),
  description: p.description ?? undefined,
  visibility: p.visibility ?? undefined,
})

const mapResource = (r: any): PermissionResource => ({
  resource_id: Number(r.resource_id),
  resource: String(r.resource ?? ''),
  permissions: Array.isArray(r.permissions) ? r.permissions.map(mapPermission) : [],
})

/** GET /org/permissions - Tenant permissions (excludes admin.*) */
export async function getPermissionResources(
  params: GetPermissionResourcesParams = {}
): Promise<PermissionResourcesResult> {
  const query = qsFrom({ page: params.page, page_size: params.page_size })
  const json = await request(`${ORG_PERMISSIONS}${query ? `?${query}` : ''}`, { method: 'GET' })
  const payload = (json as any)?.data ?? json

  const rawList = Array.isArray(payload?.data)
    ? payload.data
    : Array.isArray(payload?.items)
    ? payload.items
    : Array.isArray(payload)
    ? payload
    : Array.isArray(payload?.resources)
    ? payload.resources
    : []

  const items = rawList.map(mapResource)

  const page = Number(payload?.page ?? params.page ?? 1)
  const page_size = Number(payload?.page_size ?? params.page_size ?? (items.length || 10))
  const total = Number(payload?.total_count ?? payload?.total ?? payload?.count ?? items.length)

  return {
    items,
    total,
    page,
    page_size,
    total_pages: typeof payload?.total_pages === 'number' ? payload.total_pages : undefined,
    has_next: typeof payload?.has_next === 'boolean' ? payload.has_next : undefined,
    has_prev: typeof payload?.has_prev === 'boolean' ? payload.has_prev : undefined,
  }
}

/** GET /admin/permissions - Admin permissions (includes all) */
export async function AdmingetPermissionResources(
  params: GetPermissionResourcesParams = {}
): Promise<PermissionResourcesResult> {
  const query = qsFrom({ page: params.page, page_size: params.page_size })
  const json = await request(`${ADMIN_PERMISSIONS}${query ? `?${query}` : ''}`, { method: 'GET' })
  const payload = (json as any)?.data ?? json

  const rawList = Array.isArray(payload?.data)
    ? payload.data
    : Array.isArray(payload?.items)
    ? payload.items
    : Array.isArray(payload)
    ? payload
    : Array.isArray(payload?.resources)
    ? payload.resources
    : []

  const items = rawList.map(mapResource)

  const page = Number(payload?.page ?? params.page ?? 1)
  const page_size = Number(payload?.page_size ?? params.page_size ?? (items.length || 10))
  const total = Number(payload?.total_count ?? payload?.total ?? payload?.count ?? items.length)

  return {
    items,
    total,
    page,
    page_size,
    total_pages: typeof payload?.total_pages === 'number' ? payload.total_pages : undefined,
    has_next: typeof payload?.has_next === 'boolean' ? payload.has_next : undefined,
    has_prev: typeof payload?.has_prev === 'boolean' ? payload.has_prev : undefined,
  }
}

/** PUT /admin/permissions/visibility/prefix - Admin only */
export type SetVisibilityByPrefixInput = {
  prefix: string
  visibility: 'ALL' | 'HQ_ONLY'
}

export const setVisibilityByPrefix = (payload: SetVisibilityByPrefixInput) =>
  unwrap(
    request(`${ADMIN_PERMISSIONS}/visibility/prefix`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    })
  )

export type SetVisibilityByNameInput = {
  name: string
  visibility: 'ALL' | 'HQ_ONLY'
}

/** PUT /admin/permissions/visibility/name - Admin only */
export const setVisibilityByName = (payload: SetVisibilityByNameInput) =>
  unwrap(
    request(`${ADMIN_PERMISSIONS}/visibility/name`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    })
  )

/** GET /org/permissions-all - All tenant permissions (non-paginated) */
export async function getPermissionsAll(): Promise<PermissionResource[]> {
  const json = await request(`${ORG_PERMISSIONS}-all`, { method: 'GET' })
  const payload = (json as any)?.data ?? json

  const rawList = Array.isArray(payload?.items)
    ? payload.items
    : Array.isArray(payload?.resources)
    ? payload.resources
    : Array.isArray(payload)
    ? payload
    : []

  return rawList.map(mapResource)
}

/** GET /admin/permissions-all - All admin permissions (non-paginated) */
export async function AdmingetPermissionsAll(): Promise<PermissionResource[]> {
  const json = await request(`${ADMIN_PERMISSIONS}-all`, { method: 'GET' })
  const payload = (json as any)?.data ?? json

  const rawList = Array.isArray(payload?.items)
    ? payload.items
    : Array.isArray(payload?.resources)
    ? payload.resources
    : Array.isArray(payload)
    ? payload
    : []

  return rawList.map(mapResource)
}

export type UpdateRolePermissionsInput = {
  role_id: string
  permission_ids: number[]
}

/** PUT /admin/roles/{id}/permissions - Admin role permission assignment */
export const updateRolePermissions = (payload: UpdateRolePermissionsInput) =>
  unwrap(
    request(`${ADMIN_ROLES}/${encodeURIComponent(payload.role_id)}/permissions`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ permission_ids: payload.permission_ids }),
    })
  )

/** PUT /org/roles/{id}/permissions - Org role permission assignment */
export const updateOrgRolePermissions = (payload: UpdateRolePermissionsInput) =>
  unwrap(
    request(`${ORG_ROLES}/${encodeURIComponent(payload.role_id)}/permissions`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ permission_ids: payload.permission_ids }),
    })
  )
