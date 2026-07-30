// src/utils/users.ts
import { request, unwrap } from './common'
import { API_PREFIX } from './apiPrefix'

const BASE_ADMIN = `${API_PREFIX.ADMIN}/users`
const BASE_ORG = `${API_PREFIX.ORG}/org-users`

/** ─────────────────────
 *  Types (shared contract)
 *  ───────────────────── */
export type UserListItem = {
  id: string
  email: string
  fullname: string
  role?: string
  organization?: string
  photo_url?: string

  /** Normalized status used by UI */
  status: 'active' | 'disabled'

  /** Raw boolean some backends return (kept for debugging/migration) */
  active?: boolean

  /** If your API returns these, keep them; otherwise they'll be undefined */
  role_id?: number
  organization_id?: number

  /** Security fields */
  is_locked?: boolean
  failed_login_attempts?: number
  two_factor_enabled?: boolean
}

export type GetUsersParams = {
  page?: number
  page_size?: number
  q?: string
  organization_id?: number | string
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
  status?: 'active' | 'disabled'
  user_type?: 'system' | 'org_admin' | 'admin' | ''
}

export type UsersResult = {
  items: UserListItem[]
  total: number
  page: number
  page_size: number
  total_pages?: number
  has_next?: boolean
  has_prev?: boolean
}

/** Create payload (password optional — auto-generated if omitted) */
export type CreateUserInput = {
  fullname: string
  email: string
  password?: string
  role_id?: number
  organization_id?: number
  status: 'active' | 'disabled'
}


export type UpdateUserInput = {
  fullname: string
  email: string
  role_id?: number
  organization_id?: number
  status: 'active' | 'disabled'
}

/** ─────────────────────
 *  Helpers
 *  ───────────────────── */
const toStatus = (raw: unknown): 'active' | 'disabled' => {
  // Accepts multiple backend variants and normalizes
  if (raw === 'active') return 'active'
  if (raw === 'disabled') return 'disabled'
  if (raw === true) return 'active'
  if (raw === false) return 'disabled'
  // Fallback for null/undefined/unknown strings
  return 'active'
}

const qsFrom = (params: Record<string, unknown>) => {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return
    qs.set(k, String(v))
  })
  return qs.toString()
}

/** ─────────────────────
 *  Queries
 *  ───────────────────── */
export async function getUsers(params: GetUsersParams = {}): Promise<UsersResult> {
  const query = qsFrom({
    page: params.page,
    page_size: params.page_size,
    q: params.q,
    organization_id: params.organization_id,
    sort_by: params.sort_by,
    sort_dir: params.sort_dir,
    status: params.status,
    user_type: params.user_type,
  })

  const json = await request(`${BASE_ADMIN}${query ? `?${query}` : ''}`, { method: 'GET' })

  // Handle the actual API response structure
  const payload = (json as any)?.data ?? json
  const list = Array.isArray(payload) ? payload : Array.isArray(payload?.data) ? payload.data : []

  const items: UserListItem[] = list.map((u: any) => ({
    id: String(u.id),
    email: String(u.email ?? u.user_email ?? u.username ?? 'No email'), // Some APIs might not have email
    fullname: String(u.fullname ?? ''),
    role: typeof u.role === 'object' && u.role !== null ? u.role.name : (u.role ?? undefined),
    organization: typeof u.organization === 'object' && u.organization !== null ? u.organization.name : (u.organization ?? undefined),
    photo_url: u.photo_url ?? undefined,

    // normalize status; preserve original boolean if present
    status: toStatus(u.status ?? u.active ?? u.active_status),
    active: typeof u.active === 'boolean' ? u.active : typeof u.active_status === 'boolean' ? u.active_status : undefined,

    // keep IDs if the API provides them
    role_id: typeof u.role_id === 'number' ? u.role_id : undefined,
    organization_id: typeof u.organization_id === 'number' ? u.organization_id : undefined,

    // security fields
    is_locked: typeof u.is_locked === 'boolean' ? u.is_locked : undefined,
    failed_login_attempts: typeof u.failed_login_attempts === 'number' ? u.failed_login_attempts : undefined,
    two_factor_enabled: typeof u.two_factor_enabled === 'boolean' ? u.two_factor_enabled : undefined,
  }))

  // Extract pagination from payload (same level as data array)
  const pg = payload ?? {}
  const page = Number(pg?.page ?? params.page ?? 1)
  const page_size = Number(pg?.page_size ?? params.page_size ?? (items.length || 10))
  const total = Number(pg?.total_count ?? pg?.total ?? items.length)

  return {
    items,
    total,
    page,
    page_size,
    total_pages: typeof pg?.total_pages === 'number' ? pg.total_pages : undefined,
    has_next: typeof pg?.has_next === 'boolean' ? pg.has_next : undefined,
    has_prev: typeof pg?.has_prev === 'boolean' ? pg.has_prev : undefined,
  }
}

/** ─────────────────────
 *  Lookups (dropdowns)
 *  ───────────────────── */
export const getOrganizations = () =>
  unwrap(request(`${API_PREFIX.ADMIN}/organizations`, { method: 'GET' }))

export type UserCounts = { system: number; org_admin: number }

export const getUserCounts = (): Promise<UserCounts> =>
  unwrap(request(`${BASE_ADMIN}/counts`, { method: 'GET' }))

export const lockUserAccount = (id: string | number) =>
  unwrap(request(`${BASE_ADMIN}/${encodeURIComponent(id)}/lock`, { method: 'POST' }))

export const getRoles = () =>
  unwrap(request(`${API_PREFIX.ADMIN}/roles`, { method: 'GET' }))

/** ─────────────────────
 *  Mutations
 *  ───────────────────── */
export const createUser = (payload: CreateUserInput) =>
  unwrap(request(`${BASE_ADMIN}/create`, {
    method: 'POST',
    body: JSON.stringify(payload),
  }))

export const updateUser = (id: string | number, payload: UpdateUserInput) =>
  unwrap(request(`${BASE_ADMIN}/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  }))

export const deleteUser = (id: string | number) =>
  unwrap(request(`${BASE_ADMIN}/${encodeURIComponent(id)}`, { method: 'DELETE' }))

export const deactivateUser = (id: string | number) =>
  unwrap(request(`${BASE_ADMIN}/${encodeURIComponent(id)}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status: 'disabled', active: false }),
  }))

export const activateUser = (id: string | number) =>
  unwrap(request(`${BASE_ADMIN}/${encodeURIComponent(id)}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status: 'active', active: true }),
  }))

export const sendPasswordReset = (id: string | number) =>
  unwrap(request(`${BASE_ADMIN}/${encodeURIComponent(id)}/admin-reset-password`, {
    method: 'POST',
    body: JSON.stringify({ method: 'email', generate_password: true }),
  }))

/** Get a single org user by ID */
export async function getOrgUserById(id: string): Promise<UserListItem> {
  const json = await request(`${BASE_ORG}/${encodeURIComponent(id)}`, { method: 'GET' })
  const u = (json as any)?.data ?? json
  return {
    id: String(u.id),
    email: String(u.email ?? ''),
    fullname: String(u.fullname ?? ''),
    role: typeof u.role === 'object' && u.role !== null ? u.role.name : (u.role ?? undefined),
    organization: typeof u.organization === 'object' && u.organization !== null ? u.organization.name : (u.organization ?? undefined),
    photo_url: u.photo_url ?? undefined,
    status: toStatus(u.status ?? u.active ?? u.active_status),
    active: typeof u.active === 'boolean' ? u.active : undefined,
    role_id: typeof u.role_id === 'number' ? u.role_id : (typeof u.role === 'object' && u.role !== null ? u.role.id : undefined),
    organization_id: typeof u.organization_id === 'number' ? u.organization_id : undefined,
    is_locked: typeof u.is_locked === 'boolean' ? u.is_locked : (u.locked_at != null ? true : false),
  }
}

/** Organization Admin version - uses /api/orgs-users endpoint */
export async function getOrgUsers(params: GetUsersParams = {}): Promise<UsersResult> {
  const query = qsFrom({
    page: params.page,
    page_size: params.page_size,
    q: params.q,
    sort_by: params.sort_by,
    sort_dir: params.sort_dir,
    status: params.status,
    user_type: params.user_type,
  })

  const json = await request(`${BASE_ORG}${query ? `?${query}` : ''}`, { method: 'GET' })

  // Handle the actual API response structure
  const payload = (json as any)?.data ?? json
  const list = Array.isArray(payload) ? payload : Array.isArray(payload?.data) ? payload.data : []

  const items: UserListItem[] = list.map((u: any) => ({
    id: String(u.id),
    email: String(u.email ?? u.user_email ?? u.username ?? 'No email'), // Some APIs might not have email
    fullname: String(u.fullname ?? ''),
    role: typeof u.role === 'object' && u.role !== null ? u.role.name : (u.role ?? undefined),
    organization: typeof u.organization === 'object' && u.organization !== null ? u.organization.name : (u.organization ?? undefined),
    photo_url: u.photo_url ?? undefined,

    // normalize status; preserve original boolean if present
    status: toStatus(u.status ?? u.active ?? u.active_status),
    active: typeof u.active === 'boolean' ? u.active : typeof u.active_status === 'boolean' ? u.active_status : undefined,

    // keep IDs if the API provides them
    role_id: typeof u.role_id === 'number' ? u.role_id : undefined,
    organization_id: typeof u.organization_id === 'number' ? u.organization_id : undefined,

    // security fields
    is_locked: typeof u.is_locked === 'boolean' ? u.is_locked : undefined,
    failed_login_attempts: typeof u.failed_login_attempts === 'number' ? u.failed_login_attempts : undefined,
    two_factor_enabled: typeof u.two_factor_enabled === 'boolean' ? u.two_factor_enabled : undefined,
  }))

  // Extract pagination from payload (same level as data array)
  const pg = payload ?? {}
  const page = Number(pg?.page ?? params.page ?? 1)
  const page_size = Number(pg?.page_size ?? params.page_size ?? (items.length || 10))
  const total = Number(pg?.total_count ?? pg?.total ?? items.length)

  return {
    items,
    total,
    page,
    page_size,
    total_pages: typeof pg?.total_pages === 'number' ? pg.total_pages : undefined,
    has_next: typeof pg?.has_next === 'boolean' ? pg.has_next : undefined,
    has_prev: typeof pg?.has_prev === 'boolean' ? pg.has_prev : undefined,
  }
}

/** ─────────────────────
 *  Organization Admin Mutations
 *  ───────────────────── */
export const createOrgUser = (payload: CreateUserInput) =>
  unwrap(request(`${BASE_ORG}/create`, {
    method: 'POST',
    body: JSON.stringify(payload),
  }))

export const updateOrgUser = (id: string | number, payload: UpdateUserInput) =>
  unwrap(request(`${BASE_ORG}/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  }))

export const deleteOrgUser = (id: string | number) =>
  unwrap(request(`${BASE_ORG}/${encodeURIComponent(id)}`, { method: 'DELETE' }))

export const deactivateOrgUser = (id: string | number) =>
  unwrap(request(`${BASE_ORG}/${encodeURIComponent(id)}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status: 'disabled', active: false }),
  }))

export const activateOrgUser = (id: string | number) =>
  unwrap(request(`${BASE_ORG}/${encodeURIComponent(id)}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status: 'active', active: true }),
  }))

export const lockOrgUserAccount = (id: string | number) =>
  unwrap(request(`${BASE_ORG}/${encodeURIComponent(id)}/lock`, { method: 'POST' }))

export const unlockOrgUserAccount = (id: string | number) =>
  unwrap(request(`${BASE_ORG}/${encodeURIComponent(id)}/unlock`, { method: 'POST' }))

export const sendOrgPasswordReset = (id: string | number) =>
  unwrap(request(`${BASE_ORG}/${encodeURIComponent(id)}/reset-password`, {
    method: 'POST',
    body: JSON.stringify({ method: 'email', generate_password: true }),
  }))

/** ─────────────────────
 *  User Security Features
 *  ───────────────────── */

// Security Info Types
export type UserSecurityInfo = {
  user_id: string
  full_name: string
  email: string
  active_status: boolean
  is_locked: boolean
  locked_at?: string
  lock_reason?: string
  failed_login_attempts: number
  last_login_at?: string
  last_login_ip?: string
  two_factor_enabled: boolean
  two_factor_method?: 'email' | 'app'
  must_change_password: boolean
}

// Password Reset History Types
export type PasswordResetHistoryItem = {
  id: number
  user_id: string
  reset_by_user_id: string
  reset_by_name: string
  method: 'email' | 'form' | 'both'
  form_reference?: string
  notes?: string
  temporary_password_generated: boolean
  reset_at: string
  expiry_time?: string
  pdf_generated: boolean
  email_sent: boolean
}

// Admin Password Reset Types
export type AdminPasswordResetRequest = {
  method: 'email' | 'form' | 'both'
  notes?: string
  generate_password: boolean
  expiry_hours?: number
}

export type AdminPasswordResetResponse = {
  success: boolean
  message: string
  reset_id: number
  form_reference?: string
  temporary_password?: string
  email_sent: boolean
  pdf_generated: boolean
  pdf_url?: string
}

// Get user security info (admin only)
export const getUserSecurityInfo = (id: string | number): Promise<UserSecurityInfo> =>
  unwrap(request(`${BASE_ADMIN}/${encodeURIComponent(id)}/security`, { method: 'GET' }))

// Unlock user account
export const unlockUserAccount = (id: string | number) =>
  unwrap(request(`${BASE_ADMIN}/${encodeURIComponent(id)}/unlock`, { method: 'POST' }))

// Advanced admin password reset
export const adminResetPassword = (
  id: string | number,
  payload: AdminPasswordResetRequest
): Promise<AdminPasswordResetResponse> =>
  unwrap(request(`${BASE_ADMIN}/${encodeURIComponent(id)}/admin-reset-password`, {
    method: 'POST',
    body: JSON.stringify(payload),
  }))

// Get password reset history
export const getUserPasswordResetHistory = (
  id: string | number
): Promise<PasswordResetHistoryItem[]> =>
  unwrap(request(`${BASE_ADMIN}/${encodeURIComponent(id)}/password-reset-history`, { method: 'GET' }))

// Download password reset form PDF
export const downloadPasswordResetForm = (formId: string | number): string =>
  `${API_PREFIX.ADMIN}/password-reset-forms/${formId}/download`

/** ─────────────────────
 *  Org Admin Password Reset (mirrors admin versions but uses org routes)
 *  ───────────────────── */

// Advanced org admin password reset
export const orgAdminResetPassword = (
  id: string | number,
  payload: AdminPasswordResetRequest
): Promise<AdminPasswordResetResponse> =>
  unwrap(request(`${BASE_ORG}/${encodeURIComponent(id)}/reset-password`, {
    method: 'POST',
    body: JSON.stringify(payload),
  }))

// Get org user password reset history
export const getOrgUserPasswordResetHistory = (
  id: string | number
): Promise<PasswordResetHistoryItem[]> =>
  unwrap(request(`${BASE_ORG}/${encodeURIComponent(id)}/password-reset-history`, { method: 'GET' }))

// Download org-level password reset form PDF
export const downloadOrgPasswordResetForm = (formId: string | number): string =>
  `${API_PREFIX.ORG}/password-reset-forms/${formId}/download`

