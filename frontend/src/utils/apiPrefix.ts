/**
 * API Prefix Management
 *
 * Determines the correct API route prefix based on user permissions.
 * Note: These prefixes are appended to API_BASE (which already includes /api),
 * so the full paths become:
 * - Admin (SuperAdmin/HQ): {API_BASE}/admin/* (e.g., /api/admin/*)
 * - Organization (Tenant): {API_BASE}/org/* (e.g., /api/org/*)
 *
 * This isolation supports:
 * - Data isolation between user groups
 * - Future VPN/network-level security per route prefix
 * - Separate rate limiting and monitoring per group
 */

export const API_PREFIX = {
  ADMIN: '/admin',
  ORG: '/org',
  AUTH: '',  // Auth routes remain shared (login, logout, 2FA, etc.) - no prefix needed as API_BASE already includes /api
  PUBLIC: '/public'  // Public routes for career pages, job listings, etc. - no auth required
} as const

export type ApiPrefixType = typeof API_PREFIX[keyof typeof API_PREFIX]

/**
 * Determines API prefix based on user's permissions.
 * Maps permission prefixes to route prefixes:
 *   admin.*  -> /admin
 *   tenant.* -> /org (default)
 *
 * @param permissions - Array of permission strings from the user's role
 * @returns The appropriate API prefix for the user's access level
 */
export function getApiPrefix(permissions: string[]): ApiPrefixType {
  if (!permissions || permissions.length === 0) {
    return API_PREFIX.ORG // Default fallback
  }

  // Check for admin permissions first (highest privilege)
  if (permissions.some(p => p.startsWith('admin.'))) {
    return API_PREFIX.ADMIN
  }

  // Default to organization (tenant) routes
  return API_PREFIX.ORG
}

/**
 * Gets the user type based on permissions.
 * Useful for UI decisions and routing.
 *
 * @param permissions - Array of permission strings from the user's role
 * @returns 'admin' | 'org'
 */
export function getUserType(permissions: string[]): 'admin' | 'org' {
  if (!permissions || permissions.length === 0) {
    return 'org'
  }

  if (permissions.some(p => p.startsWith('admin.'))) {
    return 'admin'
  }

  return 'org'
}

/**
 * Builds a full API URL with the appropriate prefix based on permissions.
 *
 * @param path - The API path (e.g., '/users', '/audit/events')
 * @param permissions - Array of permission strings from the user's role
 * @returns The full API path with prefix (e.g., '/admin/users')
 */
export function buildApiPath(path: string, permissions: string[]): string {
  const prefix = getApiPrefix(permissions)
  // Ensure path starts with /
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${prefix}${normalizedPath}`
}

/**
 * Checks if the user has access to admin routes.
 *
 * @param permissions - Array of permission strings from the user's role
 * @returns true if the user has any admin.* permission
 */
export function hasAdminAccess(permissions: string[]): boolean {
  return permissions?.some(p => p.startsWith('admin.')) ?? false
}

/**
 * Checks if the user has access to organization (tenant) routes.
 *
 * @param permissions - Array of permission strings from the user's role
 * @returns true if the user has any tenant.* permission
 */
export function hasTenantAccess(permissions: string[]): boolean {
  return permissions?.some(p => p.startsWith('tenant.')) ?? false
}
