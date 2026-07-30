import React, { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { parseApiError } from '../../utils/apiError'
import {
  createAlertConfig,
  updateAlertConfig,
  type AlertConfig,
  type CreateAlertConfigInput,
  type UpdateAlertConfigInput,
  SEVERITY_LEVELS,
  EVENT_TYPES,
  CHANNEL_TYPES,
  type ChannelType,
} from '../../utils/alerting'

export type AlertConfigFormValues = {
  name: string
  enabled: boolean
  severities: string[]
  event_types: string[]
  cooldown_minutes: number
  threshold_count: number
  window_minutes: number
}

type Props = {
  initial?: AlertConfig
  mode?: 'create' | 'edit'
  onSubmit: () => Promise<void>
  onCancel: () => void
}

export default function AlertConfigForm({ initial, mode = 'create', onSubmit, onCancel }: Props) {
  const [values, setValues] = useState<AlertConfigFormValues>({
    name: initial?.name ?? '',
    enabled: initial?.enabled ?? true,
    severities: initial?.severities ?? ['WARNING', 'ERROR', 'CRITICAL'],
    event_types: initial?.event_types ?? [],
    cooldown_minutes: initial?.cooldown_seconds ? Math.round(initial.cooldown_seconds / 60) : 5,
    threshold_count: initial?.threshold_count ?? 1,
    window_minutes: initial?.window_seconds ? Math.round(initial.window_seconds / 60) : 60,
  })
  const [errors, setErrors] = useState<Partial<Record<string, string>>>({})
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (initial) {
      setValues({
        name: initial.name ?? '',
        enabled: initial.enabled ?? true,
        severities: initial.severities ?? ['WARNING', 'ERROR', 'CRITICAL'],
        event_types: initial.event_types ?? [],
        cooldown_minutes: initial.cooldown_seconds ? Math.round(initial.cooldown_seconds / 60) : 5,
        threshold_count: initial.threshold_count ?? 1,
        window_minutes: initial.window_seconds ? Math.round(initial.window_seconds / 60) : 60,
      })
    }
  }, [initial])

  const onChange = <K extends keyof AlertConfigFormValues>(k: K, v: AlertConfigFormValues[K]) => {
    setValues((s) => ({ ...s, [k]: v }))
    if (errors[k]) setErrors((e) => ({ ...e, [k]: '' }))
  }

  const toggleSeverity = (sev: string) => {
    setValues((s) => ({
      ...s,
      severities: s.severities.includes(sev)
        ? s.severities.filter((x) => x !== sev)
        : [...s.severities, sev],
    }))
  }

  const toggleEventType = (et: string) => {
    setValues((s) => ({
      ...s,
      event_types: s.event_types.includes(et)
        ? s.event_types.filter((x) => x !== et)
        : [...s.event_types, et],
    }))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setErrors({})

    // Validation
    if (!values.name.trim()) {
      setErrors({ name: 'Name is required' })
      return
    }
    if (values.severities.length === 0) {
      setErrors({ severities: 'At least one severity level is required' })
      return
    }

    setLoading(true)
    try {
      if (mode === 'create') {
        const input: CreateAlertConfigInput = {
          name: values.name.trim(),
          enabled: values.enabled,
          severities: values.severities,
          event_types: values.event_types,
          cooldown_seconds: values.cooldown_minutes * 60,
          threshold_count: values.threshold_count,
          window_seconds: values.window_minutes * 60,
        }
        await createAlertConfig(input)
      } else if (initial?.id) {
        const input: UpdateAlertConfigInput = {
          name: values.name.trim(),
          enabled: values.enabled,
          severities: values.severities,
          event_types: values.event_types,
          cooldown_seconds: values.cooldown_minutes * 60,
          threshold_count: values.threshold_count,
          window_seconds: values.window_minutes * 60,
        }
        await updateAlertConfig(initial.id, input)
      }
      await onSubmit()
    } catch (err: any) {
      const { fieldErrors, formError } = parseApiError(err)
      if (Object.keys(fieldErrors || {}).length) {
        setErrors(fieldErrors)
      }
      if (formError) toast.error(formError)
    } finally {
      setLoading(false)
    }
  }

  return (
    <form className="grid gap-5" onSubmit={handleSubmit} noValidate>
      {/* Name & Enabled */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <label className="grid gap-1">
          <span className="text-sm text-foreground/70">Configuration Name *</span>
          <input
            className={`input ${errors.name ? 'ring-1 ring-rose-400' : ''}`}
            value={values.name}
            onChange={(e) => onChange('name', e.target.value)}
            placeholder="e.g., Security Alerts"
            required
          />
          {errors.name && <span className="text-xs text-rose-400">{errors.name}</span>}
        </label>

        <label className="grid gap-1">
          <span className="text-sm text-foreground/70">Status</span>
          <div className="flex items-center gap-3 h-10">
            <button
              type="button"
              className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                values.enabled ? 'bg-brand-600' : 'bg-foreground/20'
              }`}
              onClick={() => onChange('enabled', !values.enabled)}
            >
              <span
                className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                  values.enabled ? 'translate-x-6' : 'translate-x-1'
                }`}
              />
            </button>
            <span className="text-sm">{values.enabled ? 'Enabled' : 'Disabled'}</span>
          </div>
        </label>
      </div>

      {/* Severity Levels */}
      <div className="grid gap-2">
        <span className="text-sm text-foreground/70">Severity Levels *</span>
        <div className="flex flex-wrap gap-2">
          {SEVERITY_LEVELS.map((sev) => (
            <button
              key={sev.value}
              type="button"
              onClick={() => toggleSeverity(sev.value)}
              className={`px-3 py-1.5 rounded-lg text-sm border transition-colors ${
                values.severities.includes(sev.value)
                  ? 'border-brand-500 bg-brand-500/20 text-brand-400'
                  : 'border-foreground/20 bg-foreground/5 text-foreground/60 hover:bg-foreground/10'
              }`}
            >
              <span className={sev.color}>{sev.label}</span>
            </button>
          ))}
        </div>
        {errors.severities && <span className="text-xs text-rose-400">{errors.severities}</span>}
      </div>

      {/* Event Types */}
      <div className="grid gap-2">
        <span className="text-sm text-foreground/70">Event Types (leave empty for all)</span>
        <div className="flex flex-wrap gap-2 max-h-32 overflow-y-auto p-2 bg-foreground/5 rounded-lg">
          {EVENT_TYPES.map((et) => (
            <button
              key={et.value}
              type="button"
              onClick={() => toggleEventType(et.value)}
              className={`px-2 py-1 rounded text-xs border transition-colors ${
                values.event_types.includes(et.value)
                  ? 'border-brand-500 bg-brand-500/20 text-brand-400'
                  : 'border-foreground/20 bg-foreground/5 text-foreground/60 hover:bg-foreground/10'
              }`}
            >
              {et.label}
            </button>
          ))}
        </div>
        <span className="text-xs text-foreground/40">
          {values.event_types.length === 0
            ? 'All event types will trigger this alert'
            : `${values.event_types.length} event type(s) selected`}
        </span>
      </div>

      {/* Thresholds */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <label className="grid gap-1">
          <span className="text-sm text-foreground/70">Cooldown (minutes)</span>
          <input
            type="number"
            className="input"
            value={values.cooldown_minutes}
            onChange={(e) => onChange('cooldown_minutes', Math.max(1, parseInt(e.target.value) || 1))}
            min={1}
          />
          <span className="text-xs text-foreground/40">Time between alerts</span>
        </label>

        <label className="grid gap-1">
          <span className="text-sm text-foreground/70">Threshold Count</span>
          <input
            type="number"
            className="input"
            value={values.threshold_count}
            onChange={(e) => onChange('threshold_count', Math.max(1, parseInt(e.target.value) || 1))}
            min={1}
          />
          <span className="text-xs text-foreground/40">Events before alert</span>
        </label>

        <label className="grid gap-1">
          <span className="text-sm text-foreground/70">Window (minutes)</span>
          <input
            type="number"
            className="input"
            value={values.window_minutes}
            onChange={(e) => onChange('window_minutes', Math.max(1, parseInt(e.target.value) || 1))}
            min={1}
          />
          <span className="text-xs text-foreground/40">Time window for threshold</span>
        </label>
      </div>

      {/* Actions */}
      <div className="flex justify-end gap-3 pt-2 border-t border-foreground/10">
        <button type="button" className="btn bg-foreground/10 hover:bg-foreground/20" onClick={onCancel}>
          Cancel
        </button>
        <button type="submit" className="btn" disabled={loading}>
          {loading
            ? mode === 'create'
              ? 'Creating...'
              : 'Saving...'
            : mode === 'create'
            ? 'Create Configuration'
            : 'Save Changes'}
        </button>
      </div>
    </form>
  )
}
