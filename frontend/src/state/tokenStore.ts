export type Tokens = {
  access?: string
  refresh?: string
  expiresAt?: number // timestamp when access token expires
}

const LS_KEY = 'auth_tokens'

export function saveTokens(t: Tokens) {
  const next = JSON.stringify({
    access: t.access || '',
    refresh: t.refresh || '',
    expiresAt: t.expiresAt
  })
  localStorage.setItem(LS_KEY, next)
}

export function loadTokens(): Tokens {
  try {
    const raw = localStorage.getItem(LS_KEY)
    return raw ? JSON.parse(raw) : {}
  } catch { return {} }
}

export function clearTokens() {
  localStorage.removeItem(LS_KEY)
}

// Lock state persistence (for account lockout due to failed logins)
const LOCK_KEY = 'account_locked'

export function setAccountLocked(locked: boolean) {
  if (locked) {
    localStorage.setItem(LOCK_KEY, 'true')
  } else {
    localStorage.removeItem(LOCK_KEY)
  }
}

export function isAccountLocked(): boolean {
  return localStorage.getItem(LOCK_KEY) === 'true'
}

export function clearAccountLocked() {
  localStorage.removeItem(LOCK_KEY)
}

// Session lock state persistence (for idle timeout lock screen)
const SESSION_LOCK_KEY = 'session_locked'

export function setSessionLocked(locked: boolean) {
  if (locked) {
    localStorage.setItem(SESSION_LOCK_KEY, 'true')
  } else {
    localStorage.removeItem(SESSION_LOCK_KEY)
  }
}

export function isSessionLocked(): boolean {
  return localStorage.getItem(SESSION_LOCK_KEY) === 'true'
}

export function clearSessionLocked() {
  localStorage.removeItem(SESSION_LOCK_KEY)
}

/**
 * Check if the current access token is expired
 * Returns true if expired or no expiry time is set
 */
export function isTokenExpired(): boolean {
  const tokens = loadTokens()
  if (!tokens.expiresAt) return true // No expiry time = treat as expired

  // Add 30 second buffer to refresh before actual expiry
  return Date.now() >= (tokens.expiresAt - 30000)
}

/**
 * Check if tokens exist and are valid
 */
export function hasValidTokens(): boolean {
  const tokens = loadTokens()
  return !!(tokens.access && !isTokenExpired())
}
