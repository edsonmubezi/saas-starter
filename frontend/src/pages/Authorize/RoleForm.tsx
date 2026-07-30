import React, { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { parseApiError } from '../../utils/apiError'

export type RoleFormValues = { name: string }

const FORM_FIELDS = ['name'] as const
type FormField = (typeof FORM_FIELDS)[number]
const isFormField = (x: string): x is FormField =>
  (FORM_FIELDS as readonly string[]).includes(x)

export default function RoleForm({
  initial,
  mode = 'create',
  onSubmit,
}: {
  initial?: Partial<RoleFormValues>
  mode?: 'create' | 'edit'
  onSubmit: (values: RoleFormValues) => Promise<void>
}) {
  const [values, setValues] = useState<RoleFormValues>({ name: initial?.name ?? '' })
  const [errors, setErrors] = useState<Partial<Record<FormField, string>>>({})
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setValues({ name: initial?.name ?? '' })
  }, [initial])

  function setValue<K extends keyof RoleFormValues>(key: K, val: RoleFormValues[K]) {
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

    const payload: RoleFormValues = { name: values.name.trim() }

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
        <span className="text-sm text-foreground/70">Role name</span>
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

      <div className="flex justify-end gap-2">
        <button className="btn" disabled={loading} type="submit">
          {loading ? (mode === 'create' ? 'Creating…' : 'Saving…') : mode === 'create' ? 'Create' : 'Save changes'}
        </button>
      </div>
    </form>
  )
}
