// src/utils/apiError.ts

export type BackendFieldError = { field?: string; message?: string }
export type BackendErrorShape = {
  status?: number
  message?: string
  data?: { message?: string; errors?: BackendFieldError[] }
  payload?: any
  errors?: BackendFieldError[]
}

// Backend → Frontend field name mapping
// Maps various backend formats to the frontend form field names
const FIELD_ALIASES: Record<string, string> = {
  // camelCase → snake_case/lowercase
  fullName: 'fullname',
  firstName: 'first_name',
  lastName: 'last_name',
  roleID: 'role_id',
  organizationID: 'organization_id',
  // common variations
  phonenumber: 'phone_number',
  phoneNumber: 'phone_number',
  emailAddress: 'email',
  fullname: 'fullname',
}

function normalizeFieldName(name?: string | null): string {
  if (!name) return ''

  // exact alias first (handles most common cases)
  if (FIELD_ALIASES[name]) return FIELD_ALIASES[name]

  // items[0].name → items.0.name
  const dotted = name.replace(/\[(\d+)\]/g, '.$1')

  // Handle camelCase → snake_case properly (including ID suffix)
  // "roleID" → "role_id", "organizationID" → "organization_id"
  // But "ID" alone stays "id", not "_i_d"
  let snake = dotted
    .replace(/([a-z])([A-Z])/g, '$1_$2')  // lowercase followed by uppercase
    .replace(/([A-Z]+)([A-Z][a-z])/g, '$1_$2')  // consecutive uppercase followed by lowercase
    .toLowerCase()

  // allow alias after snake just in case
  return FIELD_ALIASES[snake] || snake
}

export function parseApiError(err: unknown): {
  fieldErrors: Record<string, string>
  formError?: string
} {
  const e = err as BackendErrorShape | any

  // Try multiple paths where errors might be located
  // Backend sends: { status, message, data: { errors: [...] } }
  const list: BackendFieldError[] | undefined =
    e?.data?.errors ||           // Standard validation response
    e?.payload?.data?.errors ||  // Sometimes wrapped in payload
    e?.errors ||                 // Direct errors array
    (Array.isArray(e?.data) ? e.data : undefined)  // Sometimes data IS the array

  const fieldErrors: Record<string, string> = {}
  if (Array.isArray(list)) {
    for (const item of list) {
      const key = normalizeFieldName(item?.field)
      const msg = (item?.message || '').toString().trim()
      if (key && msg && !(key in fieldErrors)) fieldErrors[key] = msg
    }
  }

  // Extract form-level error message
  const formError =
    Object.keys(fieldErrors).length === 0
      ? (e?.message || e?.data?.message || e?.payload?.message || 'Failed to submit')
      : undefined

  return { fieldErrors, formError }
}

// Helper to get a user-friendly error message from any error
export function getErrorMessage(err: unknown): string {
  const { fieldErrors, formError } = parseApiError(err)

  if (formError) return formError

  const messages = Object.values(fieldErrors).filter(Boolean)
  if (messages.length > 0) return messages.join('. ')

  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err

  return 'An unexpected error occurred'
}
