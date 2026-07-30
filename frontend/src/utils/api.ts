import { request } from './common'
import type { Me } from './types'
import { saveTokens, clearSessionLocked } from '../state/tokenStore'
import { API_PREFIX } from './apiPrefix'

/**
 * Decode JWT payload without verification (for expiry extraction)
 * Returns null if token is invalid
 */
function decodeJWT(token: string): any {
  try {
    const base64Url = token.split('.')[1]
    if (!base64Url) return null
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch {
    return null
  }
}

export interface LoginResult {
  success: boolean
  message: string
  access_token?: string
  refresh_token?: string
  requires_2fa?: boolean
  two_factor_method?: string
  session_token?: string
  must_change_password?: boolean
}

export async function login(identifier: string, password: string): Promise<LoginResult> {
  const json = await request(`${API_PREFIX.AUTH}/login`, {
    method: 'POST',
    body: JSON.stringify({ identifier, password }),
    skipAuthIntercept: true,
  });

  const access  = json?.data?.access_token  ?? null;
  const refresh = json?.data?.refresh_token ?? null;
  const requires_2fa = json?.data?.requires_2fa ?? false;
  const two_factor_method = json?.data?.two_factor_method ?? null;
  const session_token = json?.data?.session_token ?? null;
  const must_change_password = json?.data?.must_change_password ?? false;

  if (access) {
    // Decode token to get expiry time
    const payload = decodeJWT(access)
    const expiresAt = payload?.exp ? payload.exp * 1000 : Date.now() + 3600000 // default 1hr if no exp
    saveTokens({ access, refresh, expiresAt })
    // Clear any session lock on successful login
    clearSessionLocked()
  } else if (refresh) {
    // If only refresh token, save without expiry
    saveTokens({ access: '', refresh })
    clearSessionLocked()
  }

  return {
    success: true,
    message: json?.message ?? 'Login successful',
    access_token: access,
    refresh_token: refresh,
    requires_2fa,
    two_factor_method,
    session_token,
    must_change_password,
  };
}


export async function logout() {
  // keep your existing logic if you clear tokens here
  await request(`${API_PREFIX.AUTH}/logout`, { method: 'POST' })
}


export async function getMe(): Promise<Me> {
  const json = await request(`${API_PREFIX.AUTH}/users/me`, { method: 'GET' })
  const d = json?.data?.user ?? json?.data ?? json
  const orgType = json?.data?.organization_type ?? d?.organization_type

  // Normalize email and fullname
  const email: string | null =
    d?.email ?? d?.user_email ?? d?.username ?? d?.mail ?? null

  const fullname: string | null =
    d?.fullname ??
    d?.full_name ??
    d?.name ??
    (d?.first_name && d?.last_name ? `${d.first_name} ${d.last_name}` :
      d?.first_name ?? null)

  // Flatten role.permissions -> ["organization.create", "user.view", ...]
  const permissions: string[] = Array.from(
    new Set(
      (d?.role?.permissions ?? [])
        .map((p: any) => p?.name)
        .filter(Boolean)
    )
  )

  // Map nested org (optional)
  const organization = d?.organization
    ? {
        id: String(d.organization.id),
        name: String(d.organization.name ?? ''),
        address: d.organization.address ?? '',
        contact_person: d.organization.contact_person ?? '',
        phone_number: d.organization.phone_number ?? '',
        created_at: d.organization.created_at ?? undefined,
        updated_at: d.organization.updated_at ?? undefined,
      }
    : undefined

  // Map role (optional)
  const role = d?.role
    ? { id: Number(d.role.id), name: String(d.role.name) }
    : undefined

  const me: Me = {
    id: String(d.id),
    email,
    fullname,
    active_status: d.active_status ?? undefined,
    created_at: d.created_at ?? undefined,
    updated_at: d.updated_at ?? undefined,
    photo: d.photo ?? '',
    organization_id: typeof d.organization_id === 'number'
      ? d.organization_id
      : (d.organization_id ? Number(d.organization_id) : undefined),
    must_change_password: d.must_change_password ?? false,
    email_verified: d.email_verified ?? false,
    phone_number: d.phone_number ?? null,
    organization_type: orgType as Me['organization_type'],
    organization,
    role,
    permissions,
  }

  return me
}




