import React, { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { parseApiError } from '../../utils/apiError'

export type OrgFormValues = {
  name: string
  phone_number?: string
  address?: string
  contact_person?: string
  logo_url?: string
  email?: string
  tin?: string
  registration_number?: string
  logo?: File  // Added for file upload
}

const FORM_FIELDS = ['name', 'phone_number', 'address', 'contact_person', 'logo_url', 'email', 'tin', 'registration_number'] as const
type FormField = (typeof FORM_FIELDS)[number]
const isFormField = (x: string): x is FormField => (FORM_FIELDS as readonly string[]).includes(x)

export default function OrgsForm({
  initial,
  mode = 'create',
  onSubmit,
}: {
  initial?: Partial<OrgFormValues>
  mode?: 'create' | 'edit'
  onSubmit: (values: OrgFormValues) => Promise<void>
}) {
  const [values, setValues] = useState<OrgFormValues>({
    name: initial?.name ?? '',
    phone_number: initial?.phone_number ?? '',
    address: initial?.address ?? '',
    contact_person: initial?.contact_person ?? '',
    logo_url: initial?.logo_url ?? '',
    email: initial?.email ?? '',
    tin: initial?.tin ?? '',
    registration_number: initial?.registration_number ?? '',
  })

  const [logoFile, setLogoFile] = useState<File | null>(null)
  const [logoPreview, setLogoPreview] = useState<string | null>(initial?.logo_url ?? null)

  useEffect(() => {
    setValues({
      name: initial?.name ?? '',
      phone_number: initial?.phone_number ?? '',
      address: initial?.address ?? '',
      contact_person: initial?.contact_person ?? '',
      logo_url: initial?.logo_url ?? '',
      email: initial?.email ?? '',
      tin: initial?.tin ?? '',
      registration_number: initial?.registration_number ?? '',
    })
    setLogoPreview(initial?.logo_url ?? null)
    setLogoFile(null)
  }, [initial])

  const [errors, setErrors] = useState<Partial<Record<FormField, string>>>({})
  const [loading, setLoading] = useState(false)

  function setValue<K extends keyof OrgFormValues>(key: K, val: OrgFormValues[K]) {
    setValues(v => ({ ...v, [key]: val }))
    setErrors(prev => {
      if (!prev[key as FormField]) return prev
      const { [key as FormField]: _removed, ...rest } = prev
      return rest
    })
  }

  function handleLogoChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) {
      setLogoFile(null)
      setLogoPreview(initial?.logo_url ?? null)
      return
    }

    // Validate file type
    const validTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
    if (!validTypes.includes(file.type)) {
      toast.error('Invalid file type. Please upload JPG, PNG, GIF, or WebP')
      return
    }

    // Validate file size (5MB max)
    if (file.size > 5 * 1024 * 1024) {
      toast.error('File too large. Maximum size is 5MB')
      return
    }

    setLogoFile(file)

    // Create preview URL
    const reader = new FileReader()
    reader.onloadend = () => {
      setLogoPreview(reader.result as string)
    }
    reader.readAsDataURL(file)
  }

  // Cleanup preview URL on unmount
  useEffect(() => {
    return () => {
      if (logoPreview && logoPreview.startsWith('blob:')) {
        URL.revokeObjectURL(logoPreview)
      }
    }
  }, [logoPreview])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setErrors({})
    setLoading(true)

    const payload: OrgFormValues = {
      name: values.name.trim(),
      phone_number: values.phone_number?.trim() || undefined,
      address: values.address?.trim() || undefined,
      contact_person: values.contact_person?.trim() || undefined,
      logo_url: values.logo_url?.trim() || undefined,
      email: values.email?.trim() || undefined,
      tin: values.tin?.trim() || undefined,
      registration_number: values.registration_number?.trim() || undefined,
      logo: logoFile || undefined,
    }

    try {
      await onSubmit(payload)
    } catch (err: any) {
      const { fieldErrors, formError } = parseApiError(err)
      const filtered = Object.fromEntries(
        Object.entries(fieldErrors).filter(([k]) => isFormField(k))
      ) as Partial<Record<FormField, string>>

      if (Object.keys(filtered).length) setErrors(filtered)
      if (formError) toast.error(formError)
    } finally {
      setLoading(false)
    }
  }

  return (
    <form className="grid gap-4" onSubmit={handleSubmit} noValidate>
      <label className="grid gap-1">
        <span className="text-sm text-foreground/70">Name</span>
        <input
          name="name"
          className={`input ${errors.name ? 'ring-1 ring-rose-400' : ''}`}
          value={values.name}
          onChange={e => setValue('name', e.target.value)}
          aria-invalid={!!errors.name}
          required
        />
        {errors.name && <span className="text-xs text-rose-400">{errors.name}</span>}
      </label>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <label className="grid gap-1">
          <span className="text-sm text-foreground/70">Phone number</span>
          <input
            name="phone_number"
            className={`input ${errors.phone_number ? 'ring-1 ring-rose-400' : ''}`}
            value={values.phone_number ?? ''}
            onChange={e => setValue('phone_number', e.target.value)}
            aria-invalid={!!errors.phone_number}
          />
          {errors.phone_number && (
            <span className="text-xs text-rose-400">{errors.phone_number}</span>
          )}
        </label>

        <label className="grid gap-1">
          <span className="text-sm text-foreground/70">Contact person</span>
          <input
            name="contact_person"
            className={`input ${errors.contact_person ? 'ring-1 ring-rose-400' : ''}`}
            value={values.contact_person ?? ''}
            onChange={e => setValue('contact_person', e.target.value)}
            aria-invalid={!!errors.contact_person}
          />
          {errors.contact_person && (
            <span className="text-xs text-rose-400">{errors.contact_person}</span>
          )}
        </label>
      </div>

      <label className="grid gap-1">
        <span className="text-sm text-foreground/70">Address</span>
        <input
          name="address"
          className={`input ${errors.address ? 'ring-1 ring-rose-400' : ''}`}
          value={values.address ?? ''}
          onChange={e => setValue('address', e.target.value)}
          aria-invalid={!!errors.address}
        />
        {errors.address && <span className="text-xs text-rose-400">{errors.address}</span>}
      </label>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <label className="grid gap-1">
          <span className="text-sm text-foreground/70">Email</span>
          <input
            name="email"
            type="email"
            className={`input ${errors.email ? 'ring-1 ring-rose-400' : ''}`}
            value={values.email ?? ''}
            onChange={e => setValue('email', e.target.value)}
            aria-invalid={!!errors.email}
          />
          {errors.email && <span className="text-xs text-rose-400">{errors.email}</span>}
        </label>

        <label className="grid gap-1">
          <span className="text-sm text-foreground/70">Logo Image</span>
          <input
            name="logo"
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp"
            className="input file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-foreground/10 file:text-foreground hover:file:bg-foreground/20"
            onChange={handleLogoChange}
          />
          <span className="text-xs text-foreground/50">
            Accepted formats: JPG, PNG, GIF, WebP (max 5MB)
          </span>
        </label>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <label className="grid gap-1">
          <span className="text-sm text-foreground/70">TIN (Tax ID Number)</span>
          <input
            name="tin"
            className={`input ${errors.tin ? 'ring-1 ring-rose-400' : ''}`}
            value={values.tin ?? ''}
            onChange={e => setValue('tin', e.target.value)}
            aria-invalid={!!errors.tin}
            placeholder="1234567890"
          />
          {errors.tin && <span className="text-xs text-rose-400">{errors.tin}</span>}
        </label>

        <label className="grid gap-1">
          <span className="text-sm text-foreground/70">Registration Number</span>
          <input
            name="registration_number"
            className={`input ${errors.registration_number ? 'ring-1 ring-rose-400' : ''}`}
            value={values.registration_number ?? ''}
            onChange={e => setValue('registration_number', e.target.value)}
            aria-invalid={!!errors.registration_number}
            placeholder="Business registration number"
          />
          {errors.registration_number && (
            <span className="text-xs text-rose-400">{errors.registration_number}</span>
          )}
        </label>
      </div>

      {logoPreview && (
        <div className="grid gap-1">
          <span className="text-sm text-foreground/70">Logo Preview</span>
          <div className="w-32 h-32 border border-foreground/20 rounded-lg overflow-hidden bg-foreground/5 flex items-center justify-center">
            <img
              src={logoPreview}
              alt="Organization logo"
              className="max-w-full max-h-full object-contain"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = 'none';
              }}
            />
          </div>
          {logoFile && (
            <span className="text-xs text-foreground/50">
              Selected: {logoFile.name} ({(logoFile.size / 1024).toFixed(1)} KB)
            </span>
          )}
        </div>
      )}

      <div className="flex justify-end gap-2">
        <button className="btn" disabled={loading} type="submit">
          {loading ? (mode === 'create' ? 'Creating…' : 'Saving…') : mode === 'create' ? 'Create' : 'Save changes'}
        </button>
      </div>
    </form>
  )
}
