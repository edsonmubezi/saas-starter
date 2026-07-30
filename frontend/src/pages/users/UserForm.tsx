// src/pages/users/UserForm.tsx
import React, { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { getOrganizationsdropodown } from '../../utils/orgs'
import { getSystemRoles, getOrgRoles } from '../../utils/roles'
import { parseApiError } from '../../utils/apiError'
import { useAuth } from '../../state/AuthContext' //  added

export type UserFormValues = {
  fullname: string
  email: string
  status: 'active' | 'disabled'
  password?: string
  role_id?: number
  organization_id?: number
}

const FORM_FIELDS = [
  'fullname',
  'email',
  'password',
  'role_id',
  'organization_id',
  'status',
] as const
type FormField = (typeof FORM_FIELDS)[number]
const isFormField = (x: string): x is FormField => (FORM_FIELDS as readonly string[]).includes(x)

export default function UserForm({
  initial,
  mode = 'create',
  onSubmit,
  onCancel,
  hideOrganizationSelector = false,
}: {
  initial?: Partial<UserFormValues>
  mode?: 'create' | 'edit'
  onSubmit: (values: UserFormValues) => Promise<void>
  onCancel?: () => void
  hideOrganizationSelector?: boolean
}) {
  // 🔐 permission check
  const { user } = useAuth()
  const perms: string[] = Array.isArray((user as any)?.permissions) ? (user as any).permissions : []
  const canSeeOrg = perms.includes('admin.organization.view') && !hideOrganizationSelector

  const [values, setValues] = useState<UserFormValues>({
    fullname: initial?.fullname ?? '',
    email: initial?.email ?? '',
    role_id:
      typeof initial?.role_id === 'number'
        ? initial.role_id
        : initial?.role_id
        ? Number(initial.role_id)
        : undefined,
    organization_id:
      typeof initial?.organization_id === 'number'
        ? initial.organization_id
        : initial?.organization_id
        ? Number(initial.organization_id)
        : undefined,
    status: (initial?.status as 'active' | 'disabled') ?? 'active',
    password: '',
  })

  useEffect(() => {
    setValues(v => ({
      ...v,
      fullname: initial?.fullname ?? '',
      email: initial?.email ?? '',
      role_id:
        typeof initial?.role_id === 'number'
          ? initial.role_id
          : initial?.role_id
          ? Number(initial.role_id)
          : undefined,
      organization_id:
        typeof initial?.organization_id === 'number'
          ? initial.organization_id
          : initial?.organization_id
          ? Number(initial.organization_id)
          : undefined,
      status: (initial?.status as 'active' | 'disabled') ?? 'active',
      password: '',
    }))
  }, [initial])

  const [errors, setErrors] = useState<Partial<Record<FormField, string>>>({})
  const [loading, setLoading] = useState(false)
  const [roles, setRoles] = useState<{ id: string; name: string }[]>([])
  const [orgs, setOrgs] = useState<{ id: number; name: string }[]>([])
 

  // Load roles + (conditionally) orgs
  useEffect(() => {
    ;(async () => {
      try {
        const [r, o] = await Promise.all([
          // Use org-roles for org admin context, system roles for superadmin context
          hideOrganizationSelector ? getOrgRoles() : getSystemRoles(),
          canSeeOrg ? getOrganizationsdropodown() : Promise.resolve([]),
        ])
        setRoles(
          Array.isArray(r)
            ? r
                .map(rr => ({ id: String(rr.id), name: rr.name }))
                .filter(x => !Number.isNaN(x.id))
            : []
        )

        if (canSeeOrg) {
          setOrgs(Array.isArray(o) ? o : [])
        } else {
          setOrgs([])
          // ensure no org is sent when user can't see orgs
          setValues(v => ({ ...v, organization_id: undefined }))
        }
      } catch {
        setRoles([])
        setOrgs([])
      }
    })()
  }, [canSeeOrg, hideOrganizationSelector])

  function setValue<K extends keyof UserFormValues>(key: K, val: UserFormValues[K]) {
    setValues(v => ({ ...v, [key]: val }))
    setErrors(prev => {
      if (!prev[key as FormField]) return prev
      const { [key as FormField]: _removed, ...rest } = prev
      return rest
    })
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setErrors({})
    setLoading(true)

    // Build payload; strip organization_id if user can't see orgs
    const payloadBase: UserFormValues = {
      fullname: values.fullname.trim(),
      email: values.email.trim(),
      status: values.status,
      role_id: values.role_id ?? undefined,
      organization_id: values.organization_id ?? undefined,
      password: undefined,
    }
    const payload = canSeeOrg
      ? payloadBase
      : (({ organization_id, ...rest }) => rest)(payloadBase) as UserFormValues

    try {
      await onSubmit(payload) // parent handles success toast
    } catch (err: any) {
      const { fieldErrors, formError } = parseApiError(err)
      const filtered = Object.fromEntries(
        Object.entries(fieldErrors).filter(([k]) => isFormField(k))
      ) as Partial<Record<FormField, string>>

      // If org is hidden, drop any org-related error noise
      if (!canSeeOrg) {
        delete filtered.organization_id
      }

      if (Object.keys(filtered).length) setErrors(filtered)
      if (formError) toast.error(formError)
    } finally {
      setLoading(false)
    }
  }

  return (
    <form className="grid gap-4" onSubmit={handleSubmit} noValidate>
      <label className="grid gap-1">
        <span className="text-sm text-foreground/70">Full name</span>
        <input
          name="fullname"
          className={`input ${errors.fullname ? 'ring-1 ring-rose-400' : ''}`}
          value={values.fullname}
          onChange={e => setValue('fullname', e.target.value)}
          aria-invalid={!!errors.fullname}
        />
        {errors.fullname && <span className="text-xs text-rose-400">{errors.fullname}</span>}
      </label>

      <label className="grid gap-1">
        <span className="text-sm text-foreground/70">Email</span>
        <input
          name="email"
          className={`input ${errors.email ? 'ring-1 ring-rose-400' : ''}`}
          type="email"
          value={values.email}
          onChange={e => setValue('email', e.target.value)}
          aria-invalid={!!errors.email}
        />
        {errors.email && <span className="text-xs text-rose-400">{errors.email}</span>}
      </label>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <label className="grid gap-1">
          <span className="text-sm text-foreground/70">Role</span>
          <select
            name="role_id"
            className={`select ${errors.role_id ? 'ring-1 ring-rose-400' : ''}`}
            value={values.role_id ?? ''}
            onChange={e => setValue('role_id', e.target.value === '' ? undefined : Number(e.target.value))}
            aria-invalid={!!errors.role_id}
          >
            <option value="">— Select —</option>
            {roles.map(r => (
              <option key={r.id} value={r.id}>{r.name}</option>
            ))}
          </select>
          {errors.role_id && <span className="text-xs text-rose-400">{errors.role_id}</span>}
        </label>

        {/*  Hidden entirely if user lacks organization.view */}
        {canSeeOrg && (
          <label className="grid gap-1">
            <span className="text-sm text-foreground/70">Organization</span>
            <select
              name="organization_id"
              className={`select ${errors.organization_id ? 'ring-1 ring-rose-400' : ''}`}
              value={values.organization_id ?? ''}
              onChange={e =>
                setValue('organization_id', e.target.value === '' ? undefined : Number(e.target.value))
              }
              aria-invalid={!!errors.organization_id}
            >
              <option value="">— Select —</option>
              {orgs.map(o => (
                <option key={o.id} value={o.id}>{o.name}</option>
              ))}
            </select>
            {errors.organization_id && (
              <span className="text-xs text-rose-400">{errors.organization_id}</span>
            )}
          </label>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <label className="grid gap-1">
          <span className="text-sm text-foreground/70">Status</span>
          <select
            name="status"
            className={`select ${errors.status ? 'ring-1 ring-rose-400' : ''}`}
            value={values.status}
            onChange={e => setValue('status', e.target.value as 'active' | 'disabled')}
            aria-invalid={!!errors.status}
          >
            <option value="active">Active</option>
            <option value="disabled">Disabled</option>
          </select>
          {errors.status && <span className="text-xs text-rose-400">{errors.status}</span>}
        </label>

        {mode === 'create' && (
          <div className="flex items-center rounded-lg bg-blue-500/10 border border-blue-500/20 px-3 py-2">
            <span className="text-xs text-blue-400">A secure password will be auto-generated and sent to the user's email.</span>
          </div>
        )}
      </div>

      <div className="flex justify-end gap-2">
        <button className="btn" disabled={loading} type="submit">
          {loading
            ? mode === 'create'
              ? 'Creating…'
              : 'Saving…'
            : mode === 'create'
            ? 'Create user'
            : 'Save changes'}
        </button>
      </div>
    </form>
  )
}
