// utils/roles.ts
import { request, unwrap } from './common'
import { API_PREFIX } from './apiPrefix'

// Route prefixes for different user contexts
const ADMIN_ROLES = `${API_PREFIX.ADMIN}/roles`
const ORG_ROLES = `${API_PREFIX.ORG}/roles`

export type RolePermission = {
  id: number
  name: string
  description?: string
}

export type RoleDetails = {
  id: number | string
  name: string
  organization_id?: number
  permissions: RolePermission[]
}

export type RoleListItem = {
  id: string
  name: string
  created_at?: string
  description?: string
  permission_ids?: number[]
}

export type GetRolesParams = {
  page?: number
  page_size?: number
  q?: string
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
}

export type RolesResult = {
  items: RoleListItem[]
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

const mapRole = (o: any): RoleListItem => ({
  id: String(o.id),
  name: String(o.name ?? ''),
  created_at: o.created_at ?? undefined,
  description: o.description ?? undefined,
  permission_ids: Array.isArray(o.permission_ids) ? o.permission_ids : undefined,
})

/** GET /admin/roles - Admin roles (all roles) */
export const getRoles = () =>
  unwrap<RoleListItem[]>(
    request(`${ADMIN_ROLES}`, { method: 'GET' })
  ).then(arr => {
    const mapped = Array.isArray(arr) ? arr.map(mapRole) : []
    return mapped
  })

/** GET /admin/roles - Admin roles (unencrypted) */
export const GetUncyrptedRoles = () =>
  unwrap<RoleListItem[]>(
    request(`${ADMIN_ROLES}`, { method: 'GET' })
  ).then(arr => {
    const mapped = Array.isArray(arr) ? arr.map(mapRole) : []
    return mapped
  })

/** GET /admin/roles/system - System roles (SuperAdmin, OrgsAdmin) */
export const getSystemRoles = () =>
  request<RoleListItem[]>(`${ADMIN_ROLES}/system`, { method: 'GET' })
    .then(arr => {
      const mapped = Array.isArray(arr) ? arr.map(mapRole) : []
      return mapped
    })

/** GET /org/roles - Organization roles */
export const getOrgRoles = () =>
  unwrap<RoleListItem[]>(
    request(`${ORG_ROLES}`, { method: 'GET' })
  ).then(arr => {
    const mapped = Array.isArray(arr) ? arr.map(mapRole) : []
    return mapped
  })

/** GET /org/roles-paginated - Organization roles (paginated) */
export async function getRolesOrgsPaged(params: GetRolesParams = {}): Promise<RolesResult> {
  const query = qsFrom({
    page: params.page,
    page_size: params.page_size,
    q: params.q,
    sort_by: params.sort_by,
    sort_dir: params.sort_dir,
  })
  const url = `${ORG_ROLES}-paginated${query ? `?${query}` : ''}`

  const json = await request(url, { method: 'GET' })
  const payload = (json as any)?.data ?? json

  const rawList = Array.isArray(payload)
    ? payload
    : Array.isArray(payload?.data)
    ? payload.data
    : Array.isArray(payload?.items)
    ? payload.items
    : Array.isArray(payload?.roles)
    ? payload.roles
    : []

  const items: RoleListItem[] = rawList.map(mapRole)

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

/** GET /admin/roles-paginated - Admin roles (paginated) */
export async function getRolesPaged(params: GetRolesParams = {}): Promise<RolesResult> {
  const query = qsFrom({
    page: params.page,
    page_size: params.page_size,
    q: params.q,
    sort_by: params.sort_by,
    sort_dir: params.sort_dir,
  })
  const url = `${ADMIN_ROLES}-paginated${query ? `?${query}` : ''}`

  const json = await request(url, { method: 'GET' })
  const payload = (json as any)?.data ?? json

  const rawList = Array.isArray(payload)
    ? payload
    : Array.isArray(payload?.data)
    ? payload.data
    : Array.isArray(payload?.items)
    ? payload.items
    : Array.isArray(payload?.roles)
    ? payload.roles
    : []

  const items: RoleListItem[] = rawList.map(mapRole)

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

/** GET /org/roles/{id} - Get org role by ID */
export async function getRole(id: string | number): Promise<RoleListItem> {
  const json = await request(`${ORG_ROLES}/${encodeURIComponent(id)}`, { method: 'GET' })
  const payload = (json as any)?.data ?? json
  return mapRole(payload)
}

/** GET /admin/roles/{id} - Get admin role by ID */
export async function getAdminRole(id: string | number): Promise<RoleListItem> {
  const json = await request(`${ADMIN_ROLES}/${id}`, { method: 'GET' })
  const payload = (json as any)?.data ?? json
  return mapRole(payload)
}

/** POST /org/roles - Create org role */
export const createRole = (payload: { name: string }) =>
  unwrap(request(`${ORG_ROLES}`, { method: 'POST', body: JSON.stringify(payload) }))

/** POST /admin/roles - Create admin role */
export const createAdminRole = (payload: { name: string }) =>
  unwrap(request(`${ADMIN_ROLES}`, { method: 'POST', body: JSON.stringify(payload) }))

/** PUT /org/roles/{id} - Update org role */
export const updateRole = (id: string | number, payload: { name: string }) =>
  unwrap(request(`${ORG_ROLES}/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(payload) }))

/** PUT /admin/roles/{id} - Update admin role */
export const updateAdminRole = (id: string | number, payload: { name: string }) =>
  unwrap(request(`${ADMIN_ROLES}/${id}`, { method: 'PUT', body: JSON.stringify(payload) }))

/** GET /org/roles/{id}/details - Org role details with permissions */
export async function getRoleDetails(id: string | number): Promise<RoleDetails> {
  const json = await request(`${ORG_ROLES}/${encodeURIComponent(id)}/details`, { method: 'GET' })
  return (json as any)?.data ?? (json as RoleDetails)
}

/** GET /admin/roles/{id}/details - Admin role details with permissions */
export async function getRoleHQDetails(id: string | number): Promise<RoleDetails> {
  const json = await request(`${ADMIN_ROLES}/${encodeURIComponent(id)}/details`, { method: 'GET' })
  return (json as any)?.data ?? (json as RoleDetails)
}
