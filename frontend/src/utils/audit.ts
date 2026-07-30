import { request, unwrap } from './common'
import { API_PREFIX } from './apiPrefix'

const BASE = `${API_PREFIX.ADMIN}/audit`

// Types
export interface FieldChange {
  field: string
  old_value: unknown
  new_value: unknown
}

export interface AuditEvent {
  id: string
  audit_id: string
  timestamp: string
  actor_type: string
  actor_id?: number
  actor_email?: string
  actor_name?: string
  action: string
  resource_type: string
  resource_id?: number
  resource_name?: string
  severity: string
  ip_address?: string
  user_agent?: string
  before_state?: Record<string, unknown>
  after_state?: Record<string, unknown>
  changes?: FieldChange[]
}

export interface AuditEventsResult {
  events: AuditEvent[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface AuditFilters {
  page?: number
  page_size?: number
  resource_type?: string
  resource_id?: number
  actor_id?: number
  action?: string
  severity?: string
  from?: string
  to?: string
}

// API Functions
export async function getAuditEvents(params: AuditFilters = {}): Promise<AuditEventsResult> {
  const query = new URLSearchParams()
  if (params.page) query.set('page', String(params.page))
  if (params.page_size) query.set('page_size', String(params.page_size))
  if (params.resource_type) query.set('resource_type', params.resource_type)
  if (params.resource_id) query.set('resource_id', String(params.resource_id))
  if (params.actor_id) query.set('actor_id', String(params.actor_id))
  if (params.action) query.set('action', params.action)
  if (params.severity) query.set('severity', params.severity)
  if (params.from) query.set('from', params.from)
  if (params.to) query.set('to', params.to)

  const url = `${BASE}/events${query.toString() ? `?${query}` : ''}`
  return request<AuditEventsResult>(url)
}

export async function getAuditEventById(id: string): Promise<AuditEvent> {
  return request<AuditEvent>(`${BASE}/events/${id}`)
}

export async function getResourceHistory(resourceType: string, resourceId: number): Promise<AuditEvent[]> {
  return request<AuditEvent[]>(`${BASE}/resources/${resourceType}/${resourceId}/history`)
}

export async function getUserActivity(
  userId: number,
  params: { from?: string; to?: string } = {}
): Promise<AuditEvent[]> {
  const query = new URLSearchParams()
  if (params.from) query.set('from', params.from)
  if (params.to) query.set('to', params.to)

  const url = `${BASE}/users/${userId}/activity${query.toString() ? `?${query}` : ''}`
  return request<AuditEvent[]>(url)
}

// Metadata Types
export interface AuditMetadataItem {
  value: string
  label: string
  color?: string
}

export interface AuditActionsResponse {
  actions: AuditMetadataItem[]
}

export interface AuditSeveritiesResponse {
  severities: AuditMetadataItem[]
}

export interface AuditResourceTypesResponse {
  resource_types: AuditMetadataItem[]
}

// Fetch metadata from backend
export async function getAuditActions(): Promise<AuditMetadataItem[]> {
  const response = await request<AuditActionsResponse>(`${BASE}/metadata/actions`)
  return response.actions
}

export async function getAuditSeverities(): Promise<AuditMetadataItem[]> {
  const response = await request<AuditSeveritiesResponse>(`${BASE}/metadata/severities`)
  return response.severities
}

export async function getAuditResourceTypes(): Promise<AuditMetadataItem[]> {
  const response = await request<AuditResourceTypesResponse>(`${BASE}/metadata/resource-types`)
  return response.resource_types
}
