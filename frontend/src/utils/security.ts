import { request, unwrap } from './common'
import { API_PREFIX } from './apiPrefix'

const BASE = `${API_PREFIX.ADMIN}/security`

// Types
export interface SecurityEvent {
  id: string
  tenant_id?: number
  event_type: string
  category: string
  severity: string
  description: string
  actor_id?: number
  actor_email?: string
  actor_name?: string
  ip_address?: string
  user_agent?: string
  request_id?: string
  resource_type?: string
  resource_id?: number
  metadata?: Record<string, unknown>
  threat_indicators?: string[]
  risk_score?: number
  created_at: string
}

export interface SecurityEventsResult {
  events: SecurityEvent[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface SecurityFilters {
  page?: number
  page_size?: number
  severity?: string
  category?: string
  event_type?: string
  ip_address?: string
  actor_id?: number
  from?: string
  to?: string
}

export interface SecurityDashboard {
  total_events: number
  events_by_severity: Record<string, number>
  events_by_category: Record<string, number>
  recent_threats: SecurityEvent[]
  top_ip_addresses: Array<{ ip: string; count: number }>
  failed_logins: number
  suspicious_activities: number
  hours_analyzed: number
}

// API Functions
export async function getSecurityDashboard(hours: number = 24): Promise<SecurityDashboard> {
  const res = await request<{ data: SecurityDashboard }>(`${BASE}/dashboard?hours=${hours}`)
  return res.data
}

export async function getSecurityEvents(params: SecurityFilters = {}): Promise<SecurityEventsResult> {
  const query = new URLSearchParams()
  if (params.page) query.set('page', String(params.page))
  if (params.page_size) query.set('page_size', String(params.page_size))
  if (params.severity) query.set('severity', params.severity)
  if (params.category) query.set('category', params.category)
  if (params.event_type) query.set('event_type', params.event_type)
  if (params.ip_address) query.set('ip_address', params.ip_address)
  if (params.actor_id) query.set('actor_id', String(params.actor_id))
  if (params.from) query.set('from', params.from)
  if (params.to) query.set('to', params.to)

  const url = `${BASE}/events${query.toString() ? `?${query}` : ''}`
  const res = await request<{ data: SecurityEventsResult }>(url)
  return res.data
}

export async function getSecurityEventById(id: string): Promise<SecurityEvent> {
  const res = await request<{ data: SecurityEvent }>(`${BASE}/events/${id}`)
  return res.data
}

// Constants
export const SECURITY_SEVERITIES = [
  { value: 'INFO', label: 'Info', color: 'text-foreground/60', bgColor: 'bg-foreground/10' },
  { value: 'WARNING', label: 'Warning', color: 'text-amber-400', bgColor: 'bg-amber-500/10' },
  { value: 'HIGH', label: 'High', color: 'text-orange-400', bgColor: 'bg-orange-500/10' },
  { value: 'CRITICAL', label: 'Critical', color: 'text-rose-400', bgColor: 'bg-rose-500/10' },
]

export const SECURITY_CATEGORIES = [
  { value: 'authentication', label: 'Authentication', icon: '🔐' },
  { value: 'authorization', label: 'Authorization', icon: '🛡️' },
  { value: 'data_access', label: 'Data Access', icon: '📊' },
  { value: 'anomaly', label: 'Anomaly', icon: '⚠️' },
  { value: 'system', label: 'System', icon: '⚙️' },
]

export const SECURITY_EVENT_TYPES = [
  { value: 'LOGIN_SUCCESS', label: 'Login Success', category: 'authentication' },
  { value: 'LOGIN_FAILED', label: 'Login Failed', category: 'authentication' },
  { value: 'LOGOUT', label: 'Logout', category: 'authentication' },
  { value: 'PASSWORD_CHANGE', label: 'Password Change', category: 'authentication' },
  { value: 'PASSWORD_RESET', label: 'Password Reset', category: 'authentication' },
  { value: 'TWO_FACTOR_ENABLED', label: '2FA Enabled', category: 'authentication' },
  { value: 'TWO_FACTOR_DISABLED', label: '2FA Disabled', category: 'authentication' },
  { value: 'PERMISSION_DENIED', label: 'Permission Denied', category: 'authorization' },
  { value: 'ROLE_ASSIGNED', label: 'Role Assigned', category: 'authorization' },
  { value: 'ROLE_REVOKED', label: 'Role Revoked', category: 'authorization' },
  { value: 'SENSITIVE_DATA_ACCESS', label: 'Sensitive Data Access', category: 'data_access' },
  { value: 'BULK_EXPORT', label: 'Bulk Export', category: 'data_access' },
  { value: 'UNUSUAL_ACTIVITY', label: 'Unusual Activity', category: 'anomaly' },
  { value: 'BRUTE_FORCE_ATTEMPT', label: 'Brute Force Attempt', category: 'anomaly' },
  { value: 'SUSPICIOUS_IP', label: 'Suspicious IP', category: 'anomaly' },
]
