// src/utils/url.ts
export const API_BASE = import.meta.env.VITE_API_BASE_URL || ''
export const API_ORIGIN = API_BASE.replace(/\/api\/?$/i, '') // -> http://localhost:8082

/**
 * Converts a stored file path to a full URL for file viewing/downloading.
 * Uses /api/uploads/ so the request is proxied through nginx to the backend.
 *
 * Stored paths formats:
 * - "org-2/expense-claims/filename.pdf"
 * - "org-2/employee-23/filename.pdf"
 * - "/uploads/org-2/..." (already has prefix)
 * - "/api/uploads/org-2/..." (already has api prefix)
 */
export function toFileUrl(path?: string | null): string {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path

  // Normalize the path - remove any prefix variants if present
  let cleanPath = path.replace(/\\/g, '/') // Windows backslashes
  cleanPath = cleanPath.replace(/^\/api\/uploads\//, '').replace(/^api\/uploads\//, '')
  cleanPath = cleanPath.replace(/^\/uploads\//, '').replace(/^uploads\//, '')
  cleanPath = cleanPath.replace(/^\//, '') // Remove leading slash

  // Return URL pointing to /api/uploads/ route on backend (proxied by nginx)
  return `${API_ORIGIN}/api/uploads/${cleanPath}`
}
