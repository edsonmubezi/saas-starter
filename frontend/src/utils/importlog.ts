import { request, unwrap, QsFrom } from './common'

export interface ImportSession {
  id: string
  import_type: string
  status: string
  total_rows: number
  successful: number
  failed: number
  duplicates: number
  error_file?: string
  created_by: string
  created_by_name: string
  organization_id: string
  created_at: string
}

export interface ImportFailedRecord {
  id: string
  import_session_id: string
  row_number: number
  record_name: string
  reason: string
  record_type: string
  column_name?: string
  value?: string
  resolved_at?: string
  resolution?: string
  organization_id: string
}

export interface DuplicateResolution {
  record_id: string
  action: 'update' | 'skip'
}

export interface ResolveDuplicatesResult {
  updated: number
  skipped: number
  failed: number
  errors?: string[]
}

export interface ImportSessionsResult {
  sessions: ImportSession[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export async function getImportSessions(params: {
  page?: number
  page_size?: number
  import_type?: string
}): Promise<ImportSessionsResult> {
  const qs = QsFrom(params)
  return unwrap<ImportSessionsResult>(request(`/org/import-logs?${qs}`))
}

export async function getImportFailedRecords(sessionId: string): Promise<ImportFailedRecord[]> {
  return unwrap<ImportFailedRecord[]>(request(`/org/import-logs/${encodeURIComponent(sessionId)}/records`))
}

export function importTypeLabel(type: string): string {
  // Return the type as a human-readable label — callers define their own import type strings
  return type.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase())
}

export function statusColor(status: string): string {
  switch (status) {
    case 'completed': return 'text-green-400'
    case 'partial': return 'text-yellow-400'
    case 'failed': return 'text-red-400'
    default: return 'text-white/60'
  }
}

export async function getDuplicateRecords(sessionId: string): Promise<ImportFailedRecord[]> {
  return unwrap<ImportFailedRecord[]>(request(`/org/import-logs/${encodeURIComponent(sessionId)}/duplicates`))
}

export async function resolveDuplicates(sessionId: string, resolutions: DuplicateResolution[]): Promise<ResolveDuplicatesResult> {
  return unwrap<ResolveDuplicatesResult>(request('/org/import/resolve-duplicates', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ session_id: sessionId, resolutions }),
  }))
}
