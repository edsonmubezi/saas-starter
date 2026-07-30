// src/utils/common.ts
import toast from 'react-hot-toast'

import { loadTokens, saveTokens } from '../state/tokenStore'

export const API_BASE = import.meta.env.VITE_API_BASE_URL || ''

/**
 * Build headers for API requests.
 * - Defaults to JSON for normal requests.
 * - Auth header added from token store if present.
 * NOTE: For file uploads, use `uploadRequest` which will remove any Content-Type.
 */
export function buildHeaders(init: RequestInit = {}): Headers {
  const headers = new Headers(init.headers || {})

  // Default JSON content type if none provided
  if (!headers.has('Content-Type') && !headers.has('content-type')) {
    headers.set('Content-Type', 'application/json')
  }

  // Bearer token from your app’s token store
  const { access } = loadTokens()
  if (access && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${access}`)
  }

  return headers
}

/**
 * Attempt silent token refresh.
 * Returns true if a new access token was obtained.
 * Adjust the URL/shape to match your backend.
 */
export async function tryRefresh(): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/auth/refresh`, {
      method: 'POST',
      credentials: 'include', // include cookies if refresh uses them
    })
    if (!res.ok) return false

    const data = await res.json()
    if (data?.access_token) {
      // persist the new access token
      saveTokens({ access: data.access_token })
      return true
    }
    return false
  } catch (err) {
    console.error('Token refresh failed:', err)
    return false
  }
}

/**
 * Parse JSON if possible; otherwise return a simple object with raw text.
 */
export async function parseMaybeJSON(res: Response): Promise<any> {
  const txt = await res.text()
  if (!txt) return null
  try {
    return JSON.parse(txt)
  } catch {
    return { message: txt }
  }
}

/**
 * Use ONLY for multipart/form-data uploads.
 * - Never sets Content-Type (browser provides the multipart boundary)
 * - Preserves auth header
 * - Retries once on 401 after calling tryRefresh()
 */
export async function uploadRequest<T = any>(
  path: string,
  form: FormData,
  init: RequestInit = {},
  _retried = false
): Promise<T> {
  const isAbsolute = /^https?:\/\//i.test(path)
  const url = isAbsolute ? path : `${API_BASE}${path}`

  // Start from normal headers but strip any Content-Type
  const baseHeaders = buildHeaders(init)
  baseHeaders.delete('Content-Type')
  baseHeaders.delete('content-type')

  const doFetch = (override?: RequestInit) =>
    fetch(url, {
      credentials: 'include',
      ...init,
      ...override,
      headers: override?.headers ?? baseHeaders,
      body: form, // always FormData for uploads
    })

  // First attempt
  let res = await doFetch()

  // Auto-refresh once on 401
  if (res.status === 401 && !_retried) {
    const ok = await tryRefresh()
    if (ok) {
      // rebuild headers with new token, still ensure no Content-Type
      const refreshed = buildHeaders(init)
      refreshed.delete('Content-Type')
      refreshed.delete('content-type')
      res = await doFetch({ headers: refreshed })
    }
  }

  const json = await parseMaybeJSON(res)

  if (res.ok) {
    // Optional success toast for non-GET endpoints if API returns message
    // const msg = (json as any)?.message || (json as any)?.msg
    // if (msg) toast.success(msg)
    return json as T
  }

  // Build consistent error object
  const err: any = new Error(
    (json as any)?.message ||
      (json as any)?.data?.message ||
      res.statusText ||
      'Request failed'
  )
  err.status = res.status
  err.data = (json as any)?.data ?? json ?? null
  err.payload = json
  err.url = url

  // Don’t toast for validation errors
  if (res.status === 400 || res.status === 422) {
    throw err
  }

  // Gentle toasts for other cases (only once)
  if (res.status >= 500) {
    if (!err.__toasted) {
      err.__toasted = true
      toast.error(err.message || 'Server error')
    }
  } else if (res.status === 401) {
    if (!err.__toasted) {
      err.__toasted = true
      toast.error('Please sign in again')
    }
  } else if (res.status === 403) {
    if (!err.__toasted) {
      err.__toasted = true
      toast.error('Not permitted')
    }
  } else if (res.status === 404) {
    if (!err.__toasted) {
      err.__toasted = true
      toast.error('Not found')
    }
  }

  throw err
}
