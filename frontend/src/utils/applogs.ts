import { request, unwrap } from './common'
import { API_PREFIX } from './apiPrefix'

const BASE = `${API_PREFIX.ADMIN}/logs`

// Types
export interface ApplicationLog {
  id: string
  tenant_id?: number
  level: string
  category: string
  message: string
  trace_id?: string
  request_id?: string
  user_id?: number
  metadata?: Record<string, unknown>
  error?: string
  stack_trace?: string
  duration_ms?: number
  created_at: string
}

export interface AccessLog {
  id: string
  tenant_id?: number
  method: string
  path: string
  status_code: number
  duration_ms: number
  request_id?: string
  user_id?: number
  ip_address?: string
  user_agent?: string
  request_size?: number
  response_size?: number
  error?: string
  created_at: string
}

export interface LogsResult<T> {
  logs: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface LogFilters {
  page?: number
  page_size?: number
  level?: string
  category?: string
  search?: string
  trace_id?: string
  request_id?: string
  from_date?: string
  to_date?: string
}

export interface AccessLogFilters {
  page?: number
  page_size?: number
  method?: string
  path?: string
  request_id?: string
  from_date?: string
  to_date?: string
}

export interface LogStats {
  total_logs: number
  logs_by_level: Record<string, number>
  logs_by_category: Record<string, number>
  error_rate: number
  avg_response_time_ms: number
  hours_analyzed: number
}

export interface LogLevel {
  value: string
  label: string
  description: string
}

export interface LogCategory {
  value: string
  label: string
  description: string
}

// API Functions
export async function getApplicationLogs(params: LogFilters = {}): Promise<LogsResult<ApplicationLog>> {
  const query = new URLSearchParams()
  if (params.page) query.set('page', String(params.page))
  if (params.page_size) query.set('page_size', String(params.page_size))
  if (params.level) query.set('level', params.level)
  if (params.category) query.set('category', params.category)
  if (params.search) query.set('search', params.search)
  if (params.trace_id) query.set('trace_id', params.trace_id)
  if (params.request_id) query.set('request_id', params.request_id)
  if (params.from_date) query.set('from_date', params.from_date)
  if (params.to_date) query.set('to_date', params.to_date)

  const url = `${BASE}/application${query.toString() ? `?${query}` : ''}`
  const res = await request<{ data: LogsResult<ApplicationLog> }>(url)
  return res.data
}

export async function getAccessLogs(params: AccessLogFilters = {}): Promise<LogsResult<AccessLog>> {
  const query = new URLSearchParams()
  if (params.page) query.set('page', String(params.page))
  if (params.page_size) query.set('page_size', String(params.page_size))
  if (params.method) query.set('method', params.method)
  if (params.path) query.set('path', params.path)
  if (params.request_id) query.set('request_id', params.request_id)
  if (params.from_date) query.set('from_date', params.from_date)
  if (params.to_date) query.set('to_date', params.to_date)

  const url = `${BASE}/access${query.toString() ? `?${query}` : ''}`
  const res = await request<{ data: LogsResult<AccessLog> }>(url)
  return res.data
}

export async function getLogById(id: string): Promise<ApplicationLog> {
  const res = await request<{ data: ApplicationLog }>(`${BASE}/application/${id}`)
  return res.data
}

export async function getLogStats(hours: number = 24): Promise<LogStats> {
  const res = await request<{ data: LogStats }>(`${BASE}/stats?hours=${hours}`)
  return res.data
}

export async function getLogLevels(): Promise<LogLevel[]> {
  const res = await request<{ data: { levels: LogLevel[] } }>(`${BASE}/levels`)
  return res.data.levels
}

export async function getLogCategories(): Promise<LogCategory[]> {
  const res = await request<{ data: { categories: LogCategory[] } }>(`${BASE}/categories`)
  return res.data.categories
}

// Constants
export const LOG_LEVELS = [
  { value: 'DEBUG', label: 'Debug', color: 'text-foreground/50', bgColor: 'bg-foreground/5' },
  { value: 'INFO', label: 'Info', color: 'text-blue-400', bgColor: 'bg-blue-500/10' },
  { value: 'WARN', label: 'Warning', color: 'text-amber-400', bgColor: 'bg-amber-500/10' },
  { value: 'ERROR', label: 'Error', color: 'text-rose-400', bgColor: 'bg-rose-500/10' },
  { value: 'FATAL', label: 'Fatal', color: 'text-rose-500', bgColor: 'bg-rose-600/20' },
]

export const LOG_CATEGORIES = [
  { value: 'http', label: 'HTTP', icon: '🌐' },
  { value: 'database', label: 'Database', icon: '💾' },
  { value: 'cache', label: 'Cache', icon: '⚡' },
  { value: 'auth', label: 'Authentication', icon: '🔐' },
  { value: 'business', label: 'Business', icon: '📊' },
  { value: 'scheduler', label: 'Scheduler', icon: '⏰' },
  { value: 'external', label: 'External', icon: '🔗' },
  { value: 'system', label: 'System', icon: '⚙️' },
]

export const HTTP_METHODS = [
  { value: 'GET', label: 'GET', color: 'text-emerald-400' },
  { value: 'POST', label: 'POST', color: 'text-blue-400' },
  { value: 'PUT', label: 'PUT', color: 'text-amber-400' },
  { value: 'PATCH', label: 'PATCH', color: 'text-orange-400' },
  { value: 'DELETE', label: 'DELETE', color: 'text-rose-400' },
]
