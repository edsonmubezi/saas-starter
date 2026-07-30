import { request, unwrap } from './common'
import { API_PREFIX } from './apiPrefix'

const BASE = `${API_PREFIX.ADMIN}/organizations`
const ORG_BASE = `${API_PREFIX.ORG}/organization`

/** ─────────────────────
 *  Types (shared contract)
 *  ───────────────────── */
export type OrganizationListItem = {
  id: string
  name: string
  phone_number?: string
  address?: string
  contact_person?: string
  logo_url?: string              // Company logo path
  email?: string                 // Company email
  tin?: string                   // Tax ID number
  registration_number?: string   // Business registration number
  created_at?: string // ISO string
}

export type GetOrganizationsParams = {
  page?: number
  page_size?: number
  q?: string                // free-text search: name, phone_number, address
  sort_by?: string          // e.g. "created_at" | "name"
  sort_dir?: 'asc' | 'desc'
}

export type OrganizationsResult = {
  items: OrganizationListItem[]
  total: number
  page: number
  page_size: number
  total_pages?: number
  has_next?: boolean
  has_prev?: boolean
}

/** Create / Update payloads */
export type CreateOrganizationInput = {
  name: string
  phone_number?: string
  address?: string
  contact_person?: string
  logo_url?: string
  email?: string
  tin?: string
  registration_number?: string
}

export type Organization = {
  id: number
  name: string
  status?: 'active' | 'disabled'
  [k: string]: any
}

export type UpdateOrganizationInput = {
  name: string
  phone_number?: string
  address?: string
  contact_person?: string
  logo_url?: string
  email?: string
  tin?: string
  registration_number?: string
}

export type UpdateOrganizationWithLogoInput = {
  name?: string
  phone_number?: string
  address?: string
  contact_person?: string
  email?: string
  tin?: string
  registration_number?: string
  logo?: File
}

/** ─────────────────────
 *  Helpers
 *  ───────────────────── */
const qsFrom = (params: Record<string, unknown>) => {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === null || v === '') return
    qs.set(k, String(v))
  })
  return qs.toString()
}

const mapOrg = (o: any): OrganizationListItem => ({
  id: String(o.id),
  name: String(o.name ?? ''),
  phone_number: o.phone_number ?? undefined,
  address: o.address ?? undefined,
  contact_person: o.contact_person ?? undefined,
  logo_url: o.logo_url ?? undefined,
  email: o.email ?? undefined,
  tin: o.tin ?? undefined,
  registration_number: o.registration_number ?? undefined,
  created_at: o.created_at ?? undefined,
})

/** ─────────────────────
 *  Queries (paginated list + view)
 *  ───────────────────── */

// Paginated + searchable list
export async function getOrganizations(
  params: GetOrganizationsParams = {}
): Promise<OrganizationsResult> {
  const query = qsFrom({
    page: params.page,
    page_size: params.page_size,
    q: params.q,
    sort_by: params.sort_by,
    sort_dir: params.sort_dir,
  })

  const json = await request(`${BASE}/all${query ? `?${query}` : ''}`, { method: 'GET' })

  // Support both flat payloads and nested { data: { ... } }
  const payload = (json as any)?.data ?? json

  // Support either { data: [] } or { items: [] }
  const rawList = Array.isArray(payload?.data)
    ? payload.data
    : Array.isArray(payload?.items)
    ? payload.items
    : []

  const items: OrganizationListItem[] = rawList.map(mapOrg)

  const page = Number(payload?.page ?? params.page ?? 1)
  const page_size = Number(payload?.page_size ?? params.page_size ?? (items.length || 10))
  const total = Number(
    payload?.total_count ?? payload?.total ?? payload?.count ?? items.length
  )

  return {
    items,
    total,
    page,
    page_size,
    total_pages: typeof payload?.total_pages === 'number' ? payload.total_pages : undefined,
    has_next: typeof payload?.has_next === 'boolean' ? payload.has_next : undefined,
    has_prev: typeof payload?.has_prev === 'boolean' ? payload.has_prev : undefined,
  }
 
}

// View single organization
export async function getOrganization(id: string | number): Promise<OrganizationListItem> {
  const json = await request(`${BASE}/${id}`, { method: 'GET' })
  const payload = (json as any)?.data ?? json
  return mapOrg(payload)
}

export const getOrganizationsdropodown = () =>
  unwrap<Organization[]>(request(BASE, { method: 'GET' }))

/** ─────────────────────
 *  Mutations
 *  ───────────────────── */

// Create
export const createOrganization = (payload: CreateOrganizationInput) =>
  unwrap(
    request(BASE, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  )

// Edit/Update
export const updateOrganization = (id: string | number, payload: UpdateOrganizationInput) =>
  unwrap(
    request(`${BASE}/${id}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    })
  )

// Update with logo (multipart/form-data)
export async function updateOrganizationWithLogo(
  id: string | number,
  payload: UpdateOrganizationWithLogoInput
): Promise<OrganizationListItem> {
  const { access } = await import('../state/tokenStore').then(m => m.loadTokens())
  if (!access) throw new Error('No access token')

  const { API_BASE } = await import('./common')

  const formData = new FormData()

  // Add all text fields if they exist
  if (payload.name !== undefined && payload.name !== null) {
    formData.append('name', payload.name)
  }
  if (payload.phone_number) {
    formData.append('phone_number', payload.phone_number)
  }
  if (payload.address) {
    formData.append('address', payload.address)
  }
  if (payload.contact_person) {
    formData.append('contact_person', payload.contact_person)
  }
  if (payload.email) {
    formData.append('email', payload.email)
  }
  if (payload.tin) {
    formData.append('tin', payload.tin)
  }
  if (payload.registration_number) {
    formData.append('registration_number', payload.registration_number)
  }

  // Add logo file if provided
  if (payload.logo) {
    formData.append('logo', payload.logo)
  }

  const response = await fetch(`${API_BASE}${BASE}/${id}/update-with-logo`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${access}`
    },
    body: formData
  })

  if (!response.ok) {
    const errorText = await response.text()
    let errorMessage = `Failed to update organization: ${response.statusText}`
    try {
      const errorJson = JSON.parse(errorText)
      errorMessage = errorJson.message || errorMessage
    } catch {
      // If not JSON, use the text
      if (errorText) errorMessage = errorText
    }
    throw new Error(errorMessage)
  }

  const json = await response.json()
  const data = json?.data ?? json
  return mapOrg(data)
}

/** ─────────────────────
 *  Document Branding
 *  ───────────────────── */
export type DocumentBrandingSettings = {
  show_logo: boolean
  show_org_name: boolean
  show_address: boolean
  show_contact: boolean
  show_tin: boolean
  show_reg_number: boolean
  show_footer: boolean
  footer_text: string
  show_page_numbers: boolean
  show_generated_date: boolean
  show_watermark: boolean
  watermark_text: string
  watermark_type: 'text' | 'image'
  watermark_image_path: string
  primary_color: string
  header_text_color: string
  font_family: string
  header_org_name: string
  header_address: string
  header_phone: string
  header_email: string
  header_tin: string
  footer_org_name: string
}

export const DEFAULT_BRANDING: DocumentBrandingSettings = {
  show_logo: true,
  show_org_name: true,
  show_address: true,
  show_contact: true,
  show_tin: false,
  show_reg_number: false,
  show_footer: true,
  footer_text: '',
  show_page_numbers: true,
  show_generated_date: true,
  show_watermark: false,
  watermark_text: '',
  watermark_type: 'text',
  watermark_image_path: '',
  primary_color: '#1a365d',
  header_text_color: '#FFFFFF',
  font_family: 'Arial',
  header_org_name: '',
  header_address: '',
  header_phone: '',
  header_email: '',
  header_tin: '',
  footer_org_name: '',
}

export async function getDocumentBranding(): Promise<DocumentBrandingSettings> {
  return unwrap<DocumentBrandingSettings>(
    request(`${ORG_BASE}/document-branding`, { method: 'GET' })
  )
}

export async function saveDocumentBranding(input: DocumentBrandingSettings): Promise<void> {
  return unwrap(
    request(`${ORG_BASE}/document-branding`, {
      method: 'PUT',
      body: JSON.stringify(input),
    })
  )
}

export async function uploadWatermarkImage(file: File): Promise<string> {
  const { loadTokens } = await import('../state/tokenStore')
  const { access } = await loadTokens()
  if (!access) throw new Error('No access token')
  const { API_BASE } = await import('./common')

  const formData = new FormData()
  formData.append('watermark', file)

  const response = await fetch(`${API_BASE}${ORG_BASE}/document-branding/watermark`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${access}` },
    body: formData,
  })

  if (!response.ok) {
    const text = await response.text()
    let msg = 'Upload failed'
    try { msg = JSON.parse(text).message || msg } catch {}
    throw new Error(msg)
  }

  const json = await response.json()
  return json?.data?.path || ''
}

export async function getAvailableFonts(): Promise<string[]> {
  try {
    return await unwrap<string[]>(
      request(`${ORG_BASE}/document-branding/fonts`, { method: 'GET' })
    )
  } catch {
    return ['Arial', 'Helvetica', 'Times', 'Courier']
  }
}

/**
 *  Email Branding
 */
export type EmailBrandingSettings = {
  primary_color: string
  header_text_color: string
  accent_color: string
  font_family: string
  show_logo: boolean
  footer_text: string
  sign_off: string
}

export const DEFAULT_EMAIL_BRANDING: EmailBrandingSettings = {
  primary_color: '#4F46E5',
  header_text_color: '#FFFFFF',
  accent_color: '#4F46E5',
  font_family: 'Arial, sans-serif',
  show_logo: true,
  footer_text: '',
  sign_off: '',
}

export async function getEmailBranding(): Promise<EmailBrandingSettings> {
  return unwrap<EmailBrandingSettings>(
    request(`${ORG_BASE}/email-branding`, { method: 'GET' })
  )
}

export async function saveEmailBranding(input: EmailBrandingSettings): Promise<void> {
  return unwrap(
    request(`${ORG_BASE}/email-branding`, {
      method: 'PUT',
      body: JSON.stringify(input),
    })
  )
}
