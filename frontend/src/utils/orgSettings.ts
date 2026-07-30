import { request } from './common'
import { API_PREFIX } from './apiPrefix'

const BASE = `${API_PREFIX.ADMIN}/org-settings`

// ============================================================================
// TYPES & INTERFACES
// ============================================================================

export type OrganizationType =
  | 'single_company'
  | 'multiple_company'
  | 'multi_branch'
  | 'outsourcing'

export interface OrganizationSettings {
  id: number
  organizationId: number
  organizationType: OrganizationType
  sessionLockTimeoutMinutes: number
  createdAt: string
  updatedAt: string
}

export interface CreateOrgSettingsInput {
  organizationId: number
  organizationType: OrganizationType
}

export interface UpdateOrgSettingsInput {
  id: number
  organizationType?: OrganizationType
  sessionLockTimeoutMinutes?: number
}

// ============================================================================
// DEFAULTS
// ============================================================================

export const DEFAULT_ORG_SETTINGS: Omit<OrganizationSettings, 'id' | 'organizationId' | 'createdAt' | 'updatedAt'> = {
  organizationType: 'single_company',
  sessionLockTimeoutMinutes: 15,
}

// ============================================================================
// MAPPER
// ============================================================================

const mapOrgSettings = (raw: any): OrganizationSettings => ({
  id: Number(raw.id),
  organizationId: Number(raw.organizationId ?? raw.organization_id),
  organizationType: (raw.organizationType ?? raw.organization_type ?? 'single_company') as OrganizationType,
  sessionLockTimeoutMinutes: Number(raw.sessionLockTimeoutMinutes ?? raw.session_lock_timeout_minutes ?? 15),
  createdAt: String(raw.createdAt ?? raw.created_at ?? ''),
  updatedAt: String(raw.updatedAt ?? raw.updated_at ?? ''),
})

// ============================================================================
// API FUNCTIONS
// ============================================================================

export async function getOrgSettingsByOrganization(
  organizationId: string | number
): Promise<OrganizationSettings> {
  try {
    const json = await request(`${BASE}/organization/${organizationId}`, {
      method: 'GET',
    })
    const payload = (json as any)?.data ?? json
    return mapOrgSettings(payload)
  } catch (error: any) {
    if (error.status === 404) {
      return {
        ...DEFAULT_ORG_SETTINGS,
        id: 0,
        organizationId: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      }
    }
    throw error
  }
}

export async function getOrgSettingsById(
  id: number
): Promise<OrganizationSettings> {
  const json = await request(`${BASE}/${id}`, {
    method: 'GET',
  })
  const payload = (json as any)?.data ?? json
  return mapOrgSettings(payload)
}

export async function createOrgSettings(
  organizationId: string | number,
  data: Omit<CreateOrgSettingsInput, 'organizationId'>
): Promise<OrganizationSettings> {
  const json = await request(`${BASE}/organization/${organizationId}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      organizationType: data.organizationType,
    }),
  })
  const payload = (json as any)?.data ?? json
  return mapOrgSettings(payload)
}

export async function updateOrgSettings(
  data: UpdateOrgSettingsInput
): Promise<OrganizationSettings> {
  const json = await request(BASE, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  const payload = (json as any)?.data ?? json
  return mapOrgSettings(payload)
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

export const getOrgTypeDisplayName = (type: OrganizationType): string => {
  const displayNames: Record<OrganizationType, string> = {
    single_company: 'Single Company',
    multiple_company: 'Multiple Companies',
    multi_branch: 'Multi-Branch',
    outsourcing: 'Outsourcing',
  }
  return displayNames[type] || type
}

export const getOrgTypeDescription = (type: OrganizationType): string => {
  const descriptions: Record<OrganizationType, string> = {
    single_company: 'Standard single company setup',
    multiple_company: 'Manage multiple companies under one organization',
    multi_branch: 'Company with multiple branches',
    outsourcing: 'Outsourcing company providing external services',
  }
  return descriptions[type] || ''
}
