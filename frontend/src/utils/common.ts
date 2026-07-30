import toast from 'react-hot-toast'
import { loadTokens, saveTokens, clearTokens, setAccountLocked } from '../state/tokenStore'


export const API_BASE = import.meta.env.VITE_API_BASE_URL;

// Request timeout in milliseconds (30 seconds default)
const REQUEST_TIMEOUT = 30000;

// Refresh timeout — short, so a hanging server doesn't freeze the whole app
const REFRESH_TIMEOUT = 8000;


type JsonValue = any

// ---- AUTH STATE (shared across all requests) --------------------------------

/** When true, all new requests bail out immediately instead of hitting the server */
let authExpired = false

export function isAuthExpired() { return authExpired }

function markAuthExpired() {
  if (authExpired) return // already handled
  authExpired = true
  clearTokens()
  if (!window.location.pathname.includes('/login')) {
    toast.error('Please sign in again')
    window.dispatchEvent(new CustomEvent('auth:expired'))
  }
}

/** Called by AuthContext after successful login to reset the flag */
export function resetAuthExpired() { authExpired = false }

// ---- REFRESH (single-flight) ----------------------------------------------

let refreshPromise: Promise<boolean> | null = null

/**
 * Decode JWT payload without verification (for expiry extraction)
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

async function refreshOnce(): Promise<boolean> {
  const { refresh } = loadTokens()
  if (!refresh) return false

  // Abort controller with timeout so a hanging server doesn't freeze the app
  const controller = new AbortController()
  const tid = setTimeout(() => controller.abort(), REFRESH_TIMEOUT)

  try {
    const r = await fetch(`${API_BASE}/refresh`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ refresh_token: refresh }),
      signal: controller.signal,
    })
    clearTimeout(tid)

    const txt = await r.text()
    let json: any = null
    try { json = txt ? JSON.parse(txt) : null } catch {}

    if (!r.ok) { clearTokens(); return false }

    const newAccess  = json?.data?.access_token  ?? json?.access_token
    const newRefresh = json?.data?.refresh_token ?? json?.refresh_token ?? refresh
    if (!newAccess) { clearTokens(); return false }

    // Decode token to get expiry time
    const payload = decodeJWT(newAccess)
    const expiresAt = payload?.exp ? payload.exp * 1000 : Date.now() + 3600000 // default 1hr if no exp

    saveTokens({ access: newAccess, refresh: newRefresh, expiresAt })
    return true
  } catch {
    clearTimeout(tid)
    clearTokens()
    return false
  }
}

export async function tryRefresh(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = refreshOnce().finally(() => { refreshPromise = null })
  }
  return refreshPromise
}

// ---- HELPERS ----------------------------------------------------------------

function buildHeaders(init: RequestInit) {
  const headers = new Headers(init.headers || {})
  if (!(init.body instanceof FormData) && !headers.has('content-type')) headers.set('content-type', 'application/json')

  const { access } = loadTokens()
  // Use standard capitalization; most servers are case-insensitive,
  // but this avoids any odd proxies/middleware issues.
  if (access && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${access}`)
  }
  return headers
}

async function parseMaybeJSON(res: Response) {
  const txt = await res.text()
  if (!txt) return null
  try { return JSON.parse(txt) } catch { return { message: txt } }
}

// ---- CORE REQUEST -----------------------------------------------------------

/**
 * Central fetch with:
 * - early bail-out when auth is expired (no hanging requests)
 * - one 401 refresh retry (single-flight, with timeout)
 * - structured errors (status, data, payload, url)
 * - no toast for 400/422 (validation)
 * - request timeout protection
 */
export async function request<T = JsonValue>(
  path: string,
  init: RequestInit & { skipAuthIntercept?: boolean } = {},
  _retried = false
): Promise<T> {
  const skipAuth = init.skipAuthIntercept ?? false

  // Bail out immediately if we already know auth is expired
  // (but never for auth endpoints like /login that don't need a session)
  if (authExpired && !skipAuth) {
    const err: any = new Error('Session expired')
    err.status = 401
    err.isAuthExpired = true
    throw err
  }

  const isAbsolute = /^https?:\/\//i.test(path)
  const url = isAbsolute ? path : `${API_BASE}${path}`

  const headers = buildHeaders(init)

  // Create abort controller for timeout
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), REQUEST_TIMEOUT)

  const doFetch = (override?: RequestInit) =>
    fetch(url, {
      credentials: 'include', // keep if refresh uses cookies; harmless otherwise
      ...init,
      ...override,
      headers: override?.headers ?? headers,
      signal: controller.signal,
    })

  let res: Response
  try {
    res = await doFetch()
  } catch (err: any) {
    clearTimeout(timeoutId)
    if (err.name === 'AbortError') {
      const timeoutErr: any = new Error('Request timed out. Please try again.')
      timeoutErr.status = 408
      timeoutErr.isTimeout = true
      toast.error('Request timed out')
      throw timeoutErr
    }
    throw err
  }
  clearTimeout(timeoutId)

  // Attempt auto-refresh once on 401 (skip for auth endpoints like /login)
  if (res.status === 401 && !_retried && !skipAuth) {
    const ok = await tryRefresh()
    if (ok) {
      // Create a NEW abort controller for the retry (the old timeout was cleared)
      const retryController = new AbortController()
      const retryTimeout = setTimeout(() => retryController.abort(), REQUEST_TIMEOUT)
      const h2 = buildHeaders(init) // rebuild to inject new access token
      try {
        res = await fetch(url, {
          credentials: 'include',
          ...init,
          headers: h2,
          signal: retryController.signal,
        })
      } catch (retryErr: any) {
        clearTimeout(retryTimeout)
        if (retryErr.name === 'AbortError') {
          const timeoutErr: any = new Error('Request timed out. Please try again.')
          timeoutErr.status = 408
          timeoutErr.isTimeout = true
          throw timeoutErr
        }
        throw retryErr
      }
      clearTimeout(retryTimeout)
    }
  }

  const json = await parseMaybeJSON(res)

  if (res.ok) {
    // Optional success toast for non-GET if your API returns a message
    const method = (init.method || 'GET').toString().toUpperCase()
    if (method !== 'GET') {
      const msg = (json as any)?.message || (json as any)?.msg
      if (msg) {
        // toast.success(msg) // enable if you want
      }
    }
    return json as T
  }

  // Build one consistent error object
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

  // DO NOT toast for validation errors
  if (res.status === 400 || res.status === 422) {
    throw err
  }

  // Toast for other common cases (only once per error)
  if (res.status >= 500) {
    if (!err.__toasted) { err.__toasted = true; toast.error(err.message || 'Server error') }
  } else if (res.status === 401 && !skipAuth) {
    markAuthExpired()
    err.isAuthExpired = true
  } else if (res.status === 403) {
    // Check if the account is locked or deactivated - force logout
    const message = (json as any)?.message?.toLowerCase() || ''
    if (message.includes('locked') || message.includes('deactivated')) {
      // Persist lock state to localStorage for cross-tab and refresh handling
      setAccountLocked(true)
      markAuthExpired()
      err.isAuthExpired = true
    } else {
      if (!err.__toasted) { err.__toasted = true; toast.error('Not permitted') }
    }
  } else if (res.status === 404) {
    if (!err.__toasted) { err.__toasted = true; toast.error('Not found') }
  }

  throw err
}

// ---- UTILITIES --------------------------------------------------------------

export async function unwrap<T>(p: Promise<any>): Promise<T> {
  const json = await p
  return (json?.data ?? json) as T
}

export const QsFrom = (params: Record<string, unknown>) => {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return
    qs.set(k, String(v))
  })
  return qs.toString()
}
