import React, { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams, Link } from 'react-router-dom'
import toast from 'react-hot-toast'
import {
  ArrowLeft,
  Building2,
  Settings,
  Loader2,
  Save,
  CheckCircle,
} from 'lucide-react'

import {
  getOrgSettingsByOrganization,
  updateOrgSettings,
  createOrgSettings,
  getOrgTypeDisplayName,
  type OrganizationSettings,
  type OrganizationType,
} from '../../utils/orgSettings'

import { getOrganization, type OrganizationListItem } from '../../utils/orgs'

export default function OrgSettingsPage() {
  const { id } = useParams()
  const orgId = id ?? '' // Keep as encrypted string
  const qc = useQueryClient()

  // Fetch organization details
  const { data: org, isLoading: orgLoading } = useQuery({
    queryKey: ['organization', orgId],
    queryFn: () => getOrganization(orgId),
    enabled: !!orgId,
  })

  // Fetch organization settings (uses encrypted org ID)
  const { data: orgSettings, isLoading: settingsLoading } = useQuery({
    queryKey: ['org-settings', orgId],
    queryFn: () => getOrgSettingsByOrganization(orgId),
    enabled: !!orgId,
  })

  const isLoading = orgLoading || settingsLoading

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="text-center">
          <Loader2 className="w-12 h-12 animate-spin mx-auto mb-4 text-blue-500" />
          <p className="text-foreground/60">Loading organization settings...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header with back button */}
      <div className="flex items-center gap-4">
        <Link
          to="/view-organisations"
          className="p-2 rounded-lg bg-foreground/5 hover:bg-foreground/10 border border-foreground/10 transition-all"
        >
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-foreground">Organization Settings</h1>
          <p className="text-foreground/60">{org?.name || 'Unknown Organization'}</p>
        </div>
      </div>

      {/* Hero Header */}
      <div className="bg-gradient-to-r from-blue-600/20 via-purple-600/20 to-blue-600/20 border border-blue-500/20 rounded-2xl p-8 backdrop-blur">
        <div className="flex items-start gap-6">
          {/* Logo/Avatar */}
          <div className="w-20 h-20 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center text-foreground text-3xl font-bold shadow-lg overflow-hidden">
            {org?.logo_url ? (
              <img src={org.logo_url} alt="Logo" className="w-full h-full object-contain" />
            ) : (
              org?.name?.charAt(0)?.toUpperCase() ?? '?'
            )}
          </div>

          {/* Info */}
          <div className="flex-1">
            <h2 className="text-2xl font-bold text-foreground mb-2">{org?.name ?? '—'}</h2>
            <div className="flex flex-wrap gap-4 text-sm text-foreground/70">
              <span className="flex items-center gap-2">
                <Building2 className="w-4 h-4" />
                <span>{org?.contact_person ?? '—'}</span>
              </span>
              <span className="flex items-center gap-2">
                <span>{org?.email ?? '—'}</span>
              </span>
              <span className="flex items-center gap-2">
                <span>{org?.phone_number ?? '—'}</span>
              </span>
            </div>
            {orgSettings && (
              <div className="mt-3 flex flex-wrap gap-3">
                <span className="px-3 py-1 rounded-lg bg-foreground/10 text-sm text-foreground/80">
                  {getOrgTypeDisplayName(orgSettings.organizationType as OrganizationType)}
                </span>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Organization Settings Form */}
      <OrganizationSettingsTab
        orgId={orgId}
        settings={orgSettings}
        onRefresh={() => qc.invalidateQueries({ queryKey: ['org-settings', orgId] })}
      />
    </div>
  )
}

// ============================================================================
// Organization Settings Tab
// ============================================================================

interface OrganizationSettingsTabProps {
  orgId: string
  settings?: OrganizationSettings
  onRefresh: () => void
}

function OrganizationSettingsTab({ orgId, settings, onRefresh }: OrganizationSettingsTabProps) {
  const [formData, setFormData] = useState({
    organizationType: (settings?.organizationType || 'single_company') as OrganizationType,
    sessionLockTimeoutMinutes: settings?.sessionLockTimeoutMinutes ?? 15,
  })

  const updateMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      if (settings?.id) {
        return updateOrgSettings({
          id: settings.id,
          organizationType: data.organizationType,
          sessionLockTimeoutMinutes: data.sessionLockTimeoutMinutes,
        })
      } else {
        return createOrgSettings(orgId, {
          organizationType: data.organizationType,
        })
      }
    },
    onSuccess: () => {
      toast.success('Organization settings saved')
      onRefresh()
    },
    onError: (error: any) => {
      toast.error(error?.message || 'Failed to save settings')
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updateMutation.mutate(formData)
  }

  return (
    <form onSubmit={handleSubmit}>
      <div className="bg-foreground/5 border border-foreground/10 rounded-xl p-6 backdrop-blur space-y-6">
        <div className="flex items-center gap-2 pb-4 border-b border-foreground/10">
          <Settings className="w-6 h-6 text-blue-400" />
          <h3 className="text-lg font-semibold text-foreground">Organization Configuration</h3>
        </div>

        {/* Organization Type */}
        <div className="space-y-3">
          <label className="block text-sm font-medium text-foreground/80">Organization Type</label>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
            {([
              { value: 'single_company' as const, label: 'Single Company', desc: 'Standard single company setup' },
              { value: 'multiple_company' as const, label: 'Multiple Companies', desc: 'Multiple companies under one org' },
              { value: 'multi_branch' as const, label: 'Multi-Branch', desc: 'Company with multiple branches' },
              { value: 'outsourcing' as const, label: 'Outsourcing', desc: 'Outsourcing company providing services' },
            ]).map((option) => (
              <label
                key={option.value}
                className={`relative flex flex-col p-4 rounded-lg border cursor-pointer transition-all ${
                  formData.organizationType === option.value
                    ? 'border-blue-500 bg-blue-500/10'
                    : 'border-foreground/10 bg-foreground/5 hover:bg-foreground/10'
                }`}
              >
                <input
                  type="radio"
                  name="organizationType"
                  value={option.value}
                  checked={formData.organizationType === option.value}
                  onChange={(e) => setFormData((prev) => ({ ...prev, organizationType: e.target.value as OrganizationType }))}
                  className="sr-only"
                />
                <span className="font-medium text-foreground">{option.label}</span>
                <span className="text-xs text-foreground/60 mt-1">{option.desc}</span>
                {formData.organizationType === option.value && (
                  <CheckCircle className="absolute top-2 right-2 w-5 h-5 text-blue-400" />
                )}
              </label>
            ))}
          </div>
        </div>

        {/* Session Lock Timeout */}
        <div className="space-y-2">
          <label className="block text-sm font-medium text-foreground/80">Session Lock Timeout (minutes)</label>
          <input
            type="number"
            value={formData.sessionLockTimeoutMinutes}
            onChange={(e) => setFormData((prev) => ({ ...prev, sessionLockTimeoutMinutes: Math.max(5, Math.min(60, Number(e.target.value))) }))}
            className="input w-32"
            min={5}
            max={60}
          />
          <p className="text-xs text-foreground/50">Lock screen appears after this many minutes of inactivity (5-60)</p>
        </div>

        {/* Save Button */}
        <div className="flex justify-end pt-4 border-t border-foreground/10">
          <button
            type="submit"
            disabled={updateMutation.isPending}
            className="btn flex items-center gap-2"
          >
            {updateMutation.isPending ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Save className="w-4 h-4" />
            )}
            Save Changes
          </button>
        </div>
      </div>
    </form>
  )
}
