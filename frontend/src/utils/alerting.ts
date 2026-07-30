// src/utils/alerting.ts
import { request, unwrap, QsFrom } from './common'
import { API_PREFIX } from './apiPrefix'

const BASE = `${API_PREFIX.ORG}/alerting`

/** ─────────────────────
 *  Types
 *  ───────────────────── */

export type ChannelType = 'email' | 'slack' | 'sms' | 'teams' | 'webhook'

export type AlertChannelConfig = {
  channel_type: ChannelType
  config: Record<string, string>
  enabled: boolean
}

export type AlertConfig = {
  id: string
  tenant_id: number
  name: string
  enabled: boolean
  event_types: string[]
  severities: string[]
  conditions?: Record<string, unknown>
  cooldown_seconds: number
  threshold_count: number
  window_seconds: number
  channels: AlertChannelConfig[]
  created_at: string
  updated_at: string
  created_by: number
}

export type CreateAlertConfigInput = {
  name: string
  enabled: boolean
  event_types: string[]
  severities: string[]
  conditions?: Record<string, unknown>
  cooldown_seconds?: number
  threshold_count?: number
  window_seconds?: number
  channels?: AlertChannelConfig[]
}

export type UpdateAlertConfigInput = {
  name?: string
  enabled?: boolean
  event_types?: string[]
  severities?: string[]
  conditions?: Record<string, unknown>
  cooldown_seconds?: number
  threshold_count?: number
  window_seconds?: number
}

export type AlertHistoryItem = {
  id: string
  alert_config_id: string
  tenant_id: number
  event_id: string
  event_type: string
  severity: string
  alert_title: string
  alert_message: string
  channels_notified: string[]
  delivery_results: {
    channel: string
    success: boolean
    error?: string
    sent_at: string
  }[]
  created_at: string
}

export type GetAlertConfigsParams = {
  page?: number
  page_size?: number
}

export type AlertConfigsResult = {
  items: AlertConfig[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export type AlertHistoryResult = {
  items: AlertHistoryItem[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export type ConfiguredChannelsResult = {
  channels: string[]
  enabled: boolean
}

export type TestAlertResult = {
  channel: string
  success: boolean
  error?: string
  sent_at: string
}

/** ─────────────────────
 *  Mappers
 *  ───────────────────── */

const mapAlertConfig = (d: any): AlertConfig => ({
  id: String(d.id ?? d._id ?? ''),
  tenant_id: Number(d.tenant_id ?? 0),
  name: String(d.name ?? ''),
  enabled: Boolean(d.enabled),
  event_types: Array.isArray(d.event_types) ? d.event_types : [],
  severities: Array.isArray(d.severities) ? d.severities : [],
  conditions: d.conditions ?? undefined,
  cooldown_seconds: Number(d.cooldown_seconds ?? 300),
  threshold_count: Number(d.threshold_count ?? 1),
  window_seconds: Number(d.window_seconds ?? 3600),
  channels: Array.isArray(d.channels) ? d.channels : [],
  created_at: d.created_at ?? '',
  updated_at: d.updated_at ?? '',
  created_by: Number(d.created_by ?? 0),
})

const mapAlertHistory = (d: any): AlertHistoryItem => ({
  id: String(d.id ?? d._id ?? ''),
  alert_config_id: String(d.alert_config_id ?? ''),
  tenant_id: Number(d.tenant_id ?? 0),
  event_id: String(d.event_id ?? ''),
  event_type: String(d.event_type ?? ''),
  severity: String(d.severity ?? ''),
  alert_title: String(d.alert_title ?? ''),
  alert_message: String(d.alert_message ?? ''),
  channels_notified: Array.isArray(d.channels_notified) ? d.channels_notified : [],
  delivery_results: Array.isArray(d.delivery_results) ? d.delivery_results : [],
  created_at: d.created_at ?? '',
})

/** ─────────────────────
 *  Alert Configs CRUD
 *  ───────────────────── */

export async function getAlertConfigs(
  params: GetAlertConfigsParams = {}
): Promise<AlertConfigsResult> {
  const query = QsFrom({
    page: params.page,
    page_size: params.page_size,
  })

  const json = await request(`${BASE}/configs${query ? `?${query}` : ''}`, { method: 'GET' })
  const payload = (json as any)?.data ?? json

  const rawList = Array.isArray(payload?.configs)
    ? payload.configs
    : Array.isArray(payload?.items)
    ? payload.items
    : Array.isArray(payload)
    ? payload
    : []

  const items: AlertConfig[] = rawList.map(mapAlertConfig)

  return {
    items,
    total: Number(payload?.total ?? items.length),
    page: Number(payload?.page ?? params.page ?? 1),
    page_size: Number(payload?.page_size ?? params.page_size ?? 20),
    total_pages: Number(payload?.total_pages ?? 1),
  }
}

export async function getAlertConfig(id: string): Promise<AlertConfig> {
  const json = await request(`${BASE}/configs/${id}`, { method: 'GET' })
  const payload = (json as any)?.data ?? json
  return mapAlertConfig(payload)
}

export async function createAlertConfig(input: CreateAlertConfigInput): Promise<AlertConfig> {
  const json = await request(`${BASE}/configs`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
  const payload = (json as any)?.data ?? json
  return mapAlertConfig(payload)
}

export async function updateAlertConfig(
  id: string,
  input: UpdateAlertConfigInput
): Promise<AlertConfig> {
  const json = await request(`${BASE}/configs/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
  const payload = (json as any)?.data ?? json
  return mapAlertConfig(payload)
}

export async function deleteAlertConfig(id: string): Promise<void> {
  await request(`${BASE}/configs/${id}`, { method: 'DELETE' })
}

/** ─────────────────────
 *  Alert History
 *  ───────────────────── */

export async function getAlertHistory(
  params: GetAlertConfigsParams = {}
): Promise<AlertHistoryResult> {
  const query = QsFrom({
    page: params.page,
    page_size: params.page_size,
  })

  const json = await request(`${BASE}/history${query ? `?${query}` : ''}`, { method: 'GET' })
  const payload = (json as any)?.data ?? json

  const rawList = Array.isArray(payload?.history)
    ? payload.history
    : Array.isArray(payload?.items)
    ? payload.items
    : Array.isArray(payload)
    ? payload
    : []

  const items: AlertHistoryItem[] = rawList.map(mapAlertHistory)

  return {
    items,
    total: Number(payload?.total ?? items.length),
    page: Number(payload?.page ?? params.page ?? 1),
    page_size: Number(payload?.page_size ?? params.page_size ?? 20),
    total_pages: Number(payload?.total_pages ?? 1),
  }
}

/** ─────────────────────
 *  Channels
 *  ───────────────────── */

export async function getConfiguredChannels(): Promise<ConfiguredChannelsResult> {
  const json = await request(`${BASE}/channels`, { method: 'GET' })
  const payload = (json as any)?.data ?? json
  return {
    channels: Array.isArray(payload?.channels) ? payload.channels : [],
    enabled: Boolean(payload?.enabled),
  }
}

export async function testAlertChannel(channel: ChannelType): Promise<TestAlertResult> {
  const json = await request(`${BASE}/test/${channel}`, { method: 'POST' })
  const payload = (json as any)?.data ?? json
  return {
    channel: String(payload?.channel ?? channel),
    success: Boolean(payload?.success),
    error: payload?.error,
    sent_at: payload?.sent_at ?? new Date().toISOString(),
  }
}

/** ─────────────────────
 *  Constants
 *  ───────────────────── */

export const SEVERITY_LEVELS = [
  { value: 'INFO', label: 'Info', color: 'text-blue-400' },
  { value: 'WARNING', label: 'Warning', color: 'text-yellow-400' },
  { value: 'ERROR', label: 'Error', color: 'text-orange-400' },
  { value: 'CRITICAL', label: 'Critical', color: 'text-red-400' },
] as const

export const CHANNEL_TYPES: { value: ChannelType; label: string; icon: string }[] = [
  { value: 'email', label: 'Email', icon: 'Mail' },
  { value: 'slack', label: 'Slack', icon: 'MessageSquare' },
  { value: 'sms', label: 'SMS', icon: 'Smartphone' },
  { value: 'teams', label: 'Microsoft Teams', icon: 'Users' },
  { value: 'webhook', label: 'Webhook', icon: 'Link' },
]

export const EVENT_TYPES = [
  { value: 'FAILED_LOGIN', label: 'Failed Login' },
  { value: 'ACCOUNT_LOCKED', label: 'Account Locked' },
  { value: 'PASSWORD_RESET', label: 'Password Reset' },
  { value: 'PERMISSION_DENIED', label: 'Permission Denied' },
  { value: 'SUSPICIOUS_ACTIVITY', label: 'Suspicious Activity' },
  { value: 'DATA_EXPORT', label: 'Data Export' },
  { value: 'BULK_OPERATION', label: 'Bulk Operation' },
  { value: 'CONFIG_CHANGE', label: 'Configuration Change' },
  { value: 'USER_CREATED', label: 'User Created' },
  { value: 'USER_DELETED', label: 'User Deleted' },
  { value: 'ROLE_CHANGED', label: 'Role Changed' },
] as const
