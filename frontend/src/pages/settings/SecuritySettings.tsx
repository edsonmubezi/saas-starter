import React, { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { request } from '../../utils/common'
import toast from 'react-hot-toast'
import {
  Shield, KeyRound, Mail, Smartphone, AlertTriangle,
  Check, Copy, QrCode, RefreshCw, X, ChevronRight,
  Monitor, Laptop, Tablet, Globe, Trash2, LogOut,
  Lock, User, Eye, EyeOff
} from 'lucide-react'
import Modal from '../../ui/Modal'
import { useAuth } from '../../state/AuthContext'
type PasswordPolicy = {
  min_length: number
  require_upper: boolean
  require_lower: boolean
  require_number: boolean
  require_special: boolean
}

interface TwoFactorStatus {
  enabled: boolean
  method: string | null
  role_requires_2fa: boolean
  remaining_backup_codes: number
}

interface ActiveSession {
  id: number
  device_info: string
  ip_address: string
  user_agent: string
  browser: string
  os: string
  device_type: string
  location: string
  issued_at: string
  last_used_at: string
  expires_at: string
  is_current: boolean
}

interface SessionsResponse {
  sessions: ActiveSession[]
  current_session_id: number
  total_count: number
}

interface SetupResponse {
  method: string
  secret?: string
  provisioning_uri?: string
  message: string
}

interface VerifyResponse {
  success: boolean
  backup_codes?: string[]
  message: string
}

type TwoFactorMethod = 'totp' | 'email' | 'sms'

export default function SecuritySettingsPage() {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const [setupModal, setSetupModal] = useState(false)
  const [showChangePassword, setShowChangePassword] = useState(false)
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showCurrentPw, setShowCurrentPw] = useState(false)
  const [showNewPw, setShowNewPw] = useState(false)
  const [disableModal, setDisableModal] = useState(false)
  const [backupCodesModal, setBackupCodesModal] = useState(false)
  const [selectedMethod, setSelectedMethod] = useState<TwoFactorMethod | null>(null)
  const [setupData, setSetupData] = useState<SetupResponse | null>(null)
  const [verifyCode, setVerifyCode] = useState('')
  const [disableCode, setDisableCode] = useState('')
  const [backupCodes, setBackupCodes] = useState<string[]>([])
  const [regenerateCode, setRegenerateCode] = useState('')
  const [passwordPolicy, setPasswordPolicy] = useState<PasswordPolicy | null>(null)

  // Default password policy
  useEffect(() => {
    setPasswordPolicy({
      min_length: 8,
      require_upper: true,
      require_lower: true,
      require_number: true,
      require_special: true,
    })
  }, [])

  // Fetch 2FA status
  const { data: status, isLoading } = useQuery({
    queryKey: ['2fa-status'],
    queryFn: async () => {
      const res = await request('/auth/2fa/status', { method: 'GET' })
      return res.data as TwoFactorStatus
    }
  })

  // Fetch active sessions
  const { data: sessionsData, isLoading: sessionsLoading } = useQuery({
    queryKey: ['active-sessions'],
    queryFn: async () => {
      const res = await request('/sessions', { method: 'GET' })
      return res.data as SessionsResponse
    }
  })

  // Setup mutation
  const setupMutation = useMutation({
    mutationFn: async (method: TwoFactorMethod) => {
      const res = await request('/auth/2fa/setup', {
        method: 'POST',
        body: JSON.stringify({ method })
      })
      return res.data as SetupResponse
    },
    onSuccess: (data) => {
      setSetupData(data)
      if (data.method === 'email' || data.method === 'sms') {
        toast.success(data.message)
      }
    },
    onError: (e: any) => {
      toast.error(e?.message || 'Failed to setup 2FA')
    }
  })

  // Verify mutation
  const verifyMutation = useMutation({
    mutationFn: async (code: string) => {
      const res = await request('/auth/2fa/verify', {
        method: 'POST',
        body: JSON.stringify({ code })
      })
      return res.data as VerifyResponse
    },
    onSuccess: (data) => {
      if (data.backup_codes) {
        setBackupCodes(data.backup_codes)
        setBackupCodesModal(true)
      }
      toast.success('Two-factor authentication enabled!')
      queryClient.invalidateQueries({ queryKey: ['2fa-status'] })
      setSetupModal(false)
      setSetupData(null)
      setVerifyCode('')
      setSelectedMethod(null)
    },
    onError: (e: any) => {
      toast.error(e?.message || 'Invalid verification code')
    }
  })

  // Disable mutation
  const disableMutation = useMutation({
    mutationFn: async (code: string) => {
      await request('/auth/2fa/disable', {
        method: 'POST',
        body: JSON.stringify({ code })
      })
    },
    onSuccess: () => {
      toast.success('Two-factor authentication disabled')
      queryClient.invalidateQueries({ queryKey: ['2fa-status'] })
      setDisableModal(false)
      setDisableCode('')
    },
    onError: (e: any) => {
      toast.error(e?.message || 'Invalid verification code')
    }
  })

  // Regenerate backup codes mutation
  const regenerateMutation = useMutation({
    mutationFn: async (code: string) => {
      const res = await request('/auth/2fa/backup-codes/regenerate', {
        method: 'POST',
        body: JSON.stringify({ code })
      })
      return res.data as VerifyResponse
    },
    onSuccess: (data) => {
      if (data.backup_codes) {
        setBackupCodes(data.backup_codes)
        setBackupCodesModal(true)
      }
      toast.success('New backup codes generated')
      queryClient.invalidateQueries({ queryKey: ['2fa-status'] })
      setRegenerateCode('')
    },
    onError: (e: any) => {
      toast.error(e?.message || 'Invalid verification code')
    }
  })

  // Resend code mutation
  const resendMutation = useMutation({
    mutationFn: async () => {
      await request('/auth/2fa/send-code', { method: 'POST' })
    },
    onSuccess: () => {
      toast.success('Verification code sent')
    },
    onError: (e: any) => {
      toast.error(e?.message || 'Failed to send code')
    }
  })

  // Revoke session mutation
  const revokeSessionMutation = useMutation({
    mutationFn: async (sessionId: number) => {
      await request('/sessions/revoke', {
        method: 'POST',
        body: JSON.stringify({ session_id: sessionId })
      })
    },
    onSuccess: () => {
      toast.success('Session revoked')
      queryClient.invalidateQueries({ queryKey: ['active-sessions'] })
    },
    onError: (e: any) => {
      toast.error(e?.message || 'Failed to revoke session')
    }
  })

  // Revoke all other sessions mutation
  const revokeOthersMutation = useMutation({
    mutationFn: async () => {
      await request('/sessions/revoke-others', { method: 'POST' })
    },
    onSuccess: () => {
      toast.success('All other sessions revoked')
      queryClient.invalidateQueries({ queryKey: ['active-sessions'] })
    },
    onError: (e: any) => {
      toast.error(e?.message || 'Failed to revoke sessions')
    }
  })

  // Change password mutation
  const changePasswordMutation = useMutation({
    mutationFn: async ({ current_password, new_password }: { current_password: string; new_password: string }) => {
      await request('/change-password', {
        method: 'PUT',
        body: JSON.stringify({ current_password, new_password }),
      })
    },
    onSuccess: () => {
      toast.success('Password changed successfully')
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
      setShowChangePassword(false)
    },
    onError: (e: any) => {
      toast.error(e?.message || 'Failed to change password')
    },
  })

  function handleChangePassword() {
    if (!currentPassword || !newPassword) {
      toast.error('Please fill in all fields')
      return
    }
    if (newPassword !== confirmPassword) {
      toast.error('New passwords do not match')
      return
    }
    changePasswordMutation.mutate({ current_password: currentPassword, new_password: newPassword })
  }

  function handleSetupMethod(method: TwoFactorMethod) {
    setSelectedMethod(method)
    setupMutation.mutate(method)
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text)
    toast.success('Copied to clipboard')
  }

  function getMethodIcon(method: string | null) {
    switch (method) {
      case 'totp':
        return <KeyRound className="w-5 h-5" />
      case 'email':
        return <Mail className="w-5 h-5" />
      case 'sms':
        return <Smartphone className="w-5 h-5" />
      default:
        return <Shield className="w-5 h-5" />
    }
  }

  function getMethodName(method: string | null) {
    switch (method) {
      case 'totp':
        return 'Authenticator App'
      case 'email':
        return 'Email'
      case 'sms':
        return 'SMS'
      default:
        return 'Not configured'
    }
  }

  function getDeviceIcon(deviceType: string) {
    switch (deviceType) {
      case 'mobile':
        return <Smartphone className="w-5 h-5" />
      case 'tablet':
        return <Tablet className="w-5 h-5" />
      default:
        return <Monitor className="w-5 h-5" />
    }
  }

  function formatDate(dateStr: string) {
    const date = new Date(dateStr)
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  function getRelativeTime(dateStr: string) {
    const date = new Date(dateStr)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins} min ago`
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
    return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`
  }

  if (isLoading) {
    return (
      <div className="p-6">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-foreground/10 rounded w-1/4"></div>
          <div className="h-32 bg-foreground/10 rounded"></div>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-3xl mx-auto p-6 space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl md:text-2xl font-bold text-foreground mb-2">Account Settings</h1>
        <p className="text-foreground/60">Manage your personal details, password, and security</p>
      </div>

      {/* Personal Details Card */}
      <div className="bg-foreground/5 backdrop-blur-xl rounded-2xl border border-foreground/10 p-6">
        <div className="flex items-center gap-3 mb-5">
          <div className="p-3 bg-blue-500/20 rounded-xl text-blue-400">
            <User className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-lg font-semibold text-foreground">Personal Details</h2>
            <p className="text-sm text-foreground/60">Your account information</p>
          </div>
        </div>
        <div className="grid sm:grid-cols-2 gap-4">
          <div className="p-3 bg-foreground/5 rounded-xl">
            <p className="text-xs text-foreground/50 uppercase tracking-wider mb-1">Full Name</p>
            <p className="text-sm font-medium text-foreground">{user?.fullname || '—'}</p>
          </div>
          <div className="p-3 bg-foreground/5 rounded-xl">
            <p className="text-xs text-foreground/50 uppercase tracking-wider mb-1">Email</p>
            <p className="text-sm font-medium text-foreground">{user?.email || '—'}</p>
          </div>
          <div className="p-3 bg-foreground/5 rounded-xl">
            <p className="text-xs text-foreground/50 uppercase tracking-wider mb-1">Role</p>
            <p className="text-sm font-medium text-foreground">{user?.role?.name || '—'}</p>
          </div>
          <div className="p-3 bg-foreground/5 rounded-xl">
            <p className="text-xs text-foreground/50 uppercase tracking-wider mb-1">Organization</p>
            <p className="text-sm font-medium text-foreground">{user?.organization?.name || '—'}</p>
          </div>
        </div>
      </div>

      {/* Change Password Card */}
      <div className="bg-foreground/5 backdrop-blur-xl rounded-2xl border border-foreground/10 p-6">
        <div className="flex items-start justify-between mb-5">
          <div className="flex items-center gap-3">
            <div className="p-3 bg-amber-500/20 rounded-xl text-amber-400">
              <Lock className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-foreground">Change Password</h2>
              <p className="text-sm text-foreground/60">Update your account password</p>
            </div>
          </div>
          {!showChangePassword && (
            <button
              onClick={() => setShowChangePassword(true)}
              className="px-4 py-2 bg-foreground/10 hover:bg-foreground/15 text-foreground/80 text-sm font-medium rounded-xl transition-colors"
            >
              Change
            </button>
          )}
        </div>

        {showChangePassword && (() => {
          const rules: { label: string; met: boolean }[] = []
          if (passwordPolicy) {
            if (passwordPolicy.min_length > 0)
              rules.push({ label: `${passwordPolicy.min_length}+ characters`, met: newPassword.length >= passwordPolicy.min_length })
            if (passwordPolicy.require_upper)
              rules.push({ label: 'One uppercase letter', met: /[A-Z]/.test(newPassword) })
            if (passwordPolicy.require_lower)
              rules.push({ label: 'One lowercase letter', met: /[a-z]/.test(newPassword) })
            if (passwordPolicy.require_number)
              rules.push({ label: 'One number', met: /[0-9]/.test(newPassword) })
            if (passwordPolicy.require_special)
              rules.push({ label: 'One special character', met: /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(newPassword) })
          }
          const allMet = rules.length > 0 && rules.every(r => r.met)
          const passwordsMatch = newPassword === confirmPassword

          return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-foreground/70 mb-1.5">Current Password</label>
              <div className="relative">
                <input
                  type={showCurrentPw ? 'text' : 'password'}
                  value={currentPassword}
                  onChange={(e) => setCurrentPassword(e.target.value)}
                  className="w-full px-4 py-2.5 bg-foreground/10 border border-foreground/15 rounded-xl text-foreground placeholder-foreground/30 focus:outline-none focus:ring-2 focus:ring-blue-500/50 pr-10"
                  placeholder="Enter current password"
                />
                <button
                  type="button"
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-foreground/40 hover:text-foreground/70"
                  onClick={() => setShowCurrentPw(!showCurrentPw)}
                >
                  {showCurrentPw ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground/70 mb-1.5">New Password</label>
              <div className="relative">
                <input
                  type={showNewPw ? 'text' : 'password'}
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="w-full px-4 py-2.5 bg-foreground/10 border border-foreground/15 rounded-xl text-foreground placeholder-foreground/30 focus:outline-none focus:ring-2 focus:ring-blue-500/50 pr-10"
                  placeholder="Enter new password"
                />
                <button
                  type="button"
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-foreground/40 hover:text-foreground/70"
                  onClick={() => setShowNewPw(!showNewPw)}
                >
                  {showNewPw ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
            </div>

            {/* Password requirements */}
            {rules.length > 0 && (
              <div className="p-3 rounded-xl bg-foreground/5 border border-foreground/10">
                <p className="text-xs font-medium text-foreground/50 mb-2">Password must contain:</p>
                <div className="grid grid-cols-2 gap-x-4 gap-y-1.5">
                  {rules.map((rule) => (
                    <div key={rule.label} className="flex items-center gap-2 text-xs">
                      {!newPassword
                        ? <span className="w-3.5 h-3.5 rounded-full border border-foreground/30 flex-shrink-0" />
                        : rule.met
                          ? <Check className="w-3.5 h-3.5 text-green-400 flex-shrink-0" />
                          : <span className="w-3.5 h-3.5 flex items-center justify-center text-red-400 flex-shrink-0 font-bold">&times;</span>
                      }
                      <span className={!newPassword ? 'text-foreground/50' : rule.met ? 'text-green-400' : 'text-red-400'}>{rule.label}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div>
              <label className="block text-sm font-medium text-foreground/70 mb-1.5">Confirm New Password</label>
              <div className="relative">
                <input
                  type={showNewPw ? 'text' : 'password'}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="w-full px-4 py-2.5 bg-foreground/10 border border-foreground/15 rounded-xl text-foreground placeholder-foreground/30 focus:outline-none focus:ring-2 focus:ring-blue-500/50 pr-10"
                  placeholder="Confirm new password"
                  onKeyDown={(e) => e.key === 'Enter' && handleChangePassword()}
                />
                <button
                  type="button"
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-foreground/40 hover:text-foreground/70"
                  onClick={() => setShowNewPw(!showNewPw)}
                >
                  {showNewPw ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
            </div>

            {newPassword && confirmPassword && !passwordsMatch && (
              <div className="flex items-center gap-2 text-xs">
                <span className="w-3.5 h-3.5 flex items-center justify-center text-red-400 flex-shrink-0 font-bold">&times;</span>
                <span className="text-red-400/80">Passwords do not match</span>
              </div>
            )}

            <div className="flex gap-3 pt-1">
              <button
                onClick={() => {
                  setShowChangePassword(false)
                  setCurrentPassword('')
                  setNewPassword('')
                  setConfirmPassword('')
                }}
                className="flex-1 py-2.5 px-4 bg-foreground/10 hover:bg-foreground/15 text-foreground font-medium rounded-xl transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleChangePassword}
                disabled={changePasswordMutation.isPending || !currentPassword || !allMet || !confirmPassword || !passwordsMatch}
                className="flex-1 py-2.5 px-4 bg-gradient-to-r from-amber-500 to-orange-500 hover:brightness-110 disabled:opacity-50 text-white font-semibold rounded-xl transition-all"
              >
                {changePasswordMutation.isPending ? 'Changing...' : 'Update Password'}
              </button>
            </div>
          </div>
          )
        })()}
      </div>

      {/* 2FA Status Card */}
      <div className="bg-foreground/5 backdrop-blur-xl rounded-2xl border border-foreground/10 p-6">
        <div className="flex items-start justify-between mb-6">
          <div className="flex items-center gap-3">
            <div className={`p-3 rounded-xl ${status?.enabled ? 'bg-green-500/20 text-green-400' : 'bg-foreground/10 text-foreground/60'}`}>
              <Shield className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-foreground">Two-Factor Authentication</h2>
              <p className="text-sm text-foreground/60">
                {status?.enabled ? 'Your account is protected with 2FA' : 'Add an extra layer of security'}
              </p>
            </div>
          </div>
          <div className={`px-3 py-1 rounded-full text-sm font-medium ${
            status?.enabled ? 'bg-green-500/20 text-green-400' : 'bg-foreground/10 text-foreground/60'
          }`}>
            {status?.enabled ? 'Enabled' : 'Disabled'}
          </div>
        </div>

        {/* Role requirement warning */}
        {status?.role_requires_2fa && !status?.enabled && (
          <div className="mb-6 p-4 bg-yellow-500/10 border border-yellow-500/20 rounded-xl flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-yellow-400 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-sm text-yellow-400 font-medium">Two-factor authentication required</p>
              <p className="text-sm text-yellow-400/80">Your role requires 2FA to be enabled for security compliance.</p>
            </div>
          </div>
        )}

        {/* Current method info */}
        {status?.enabled && (
          <div className="mb-6 p-4 bg-foreground/5 rounded-xl">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-foreground/10 rounded-lg">
                  {getMethodIcon(status.method)}
                </div>
                <div>
                  <p className="text-sm text-foreground/60">Current method</p>
                  <p className="text-foreground font-medium">{getMethodName(status.method)}</p>
                </div>
              </div>
              {status.remaining_backup_codes !== undefined && (
                <div className="text-right">
                  <p className="text-sm text-foreground/60">Backup codes remaining</p>
                  <p className={`font-medium ${status.remaining_backup_codes <= 2 ? 'text-yellow-400' : 'text-foreground'}`}>
                    {status.remaining_backup_codes}
                  </p>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Action buttons */}
        <div className="flex gap-3">
          {!status?.enabled ? (
            <button
              onClick={() => setSetupModal(true)}
              className="flex-1 py-3 px-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white font-semibold rounded-xl transition-all duration-200"
            >
              Enable Two-Factor Authentication
            </button>
          ) : (
            <>
              <button
                onClick={() => {
                  setBackupCodes([])
                  setRegenerateCode('')
                  setBackupCodesModal(true)
                }}
                className="flex-1 py-3 px-4 bg-foreground/10 hover:bg-foreground/20 text-foreground font-medium rounded-xl transition-colors flex items-center justify-center gap-2"
              >
                <RefreshCw size={18} />
                Regenerate Backup Codes
              </button>
              <button
                onClick={() => setDisableModal(true)}
                className="py-3 px-4 bg-red-500/20 hover:bg-red-500/30 text-red-400 font-medium rounded-xl transition-colors"
              >
                Disable
              </button>
            </>
          )}
        </div>
      </div>

      {/* Active Sessions Card */}
      <div className="bg-foreground/5 backdrop-blur-xl rounded-2xl border border-foreground/10 p-6">
        <div className="flex items-start justify-between mb-6">
          <div className="flex items-center gap-3">
            <div className="p-3 bg-foreground/10 rounded-xl text-foreground/60">
              <Globe className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-foreground">Active Sessions</h2>
              <p className="text-sm text-foreground/60">
                Manage your active login sessions across devices
              </p>
            </div>
          </div>
          {sessionsData && sessionsData.sessions.length > 1 && (
            <button
              onClick={() => revokeOthersMutation.mutate()}
              disabled={revokeOthersMutation.isPending}
              className="px-3 py-1.5 bg-red-500/20 hover:bg-red-500/30 text-red-400 text-sm font-medium rounded-lg transition-colors flex items-center gap-2"
            >
              <LogOut size={14} />
              Sign out all other sessions
            </button>
          )}
        </div>

        {sessionsLoading ? (
          <div className="space-y-3">
            {[1, 2].map((i) => (
              <div key={i} className="animate-pulse p-4 bg-foreground/5 rounded-xl">
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 bg-foreground/10 rounded-lg"></div>
                  <div className="flex-1 space-y-2">
                    <div className="h-4 bg-foreground/10 rounded w-1/3"></div>
                    <div className="h-3 bg-foreground/10 rounded w-1/2"></div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : sessionsData?.sessions.length === 0 ? (
          <div className="text-center py-8 text-foreground/60">
            No active sessions found
          </div>
        ) : (
          <div className="space-y-3">
            {sessionsData?.sessions.map((session) => (
              <div
                key={session.id}
                className={`p-4 rounded-xl border transition-all ${
                  session.is_current
                    ? 'bg-blue-500/10 border-blue-500/30'
                    : 'bg-foreground/5 border-foreground/10 hover:border-foreground/20'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-4">
                    <div className={`p-2.5 rounded-lg ${
                      session.is_current ? 'bg-blue-500/20 text-blue-400' : 'bg-foreground/10 text-foreground/60'
                    }`}>
                      {getDeviceIcon(session.device_type)}
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-foreground">
                          {session.browser} on {session.os}
                        </span>
                        {session.is_current && (
                          <span className="px-2 py-0.5 bg-blue-500/20 text-blue-400 text-xs font-medium rounded-full">
                            Current
                          </span>
                        )}
                      </div>
                      <div className="flex items-center gap-3 mt-1 text-sm text-foreground/60">
                        <span className="flex items-center gap-1">
                          <Globe size={12} />
                          {session.ip_address || 'Unknown IP'}
                        </span>
                        <span>•</span>
                        <span>Last active {getRelativeTime(session.last_used_at)}</span>
                      </div>
                      <div className="text-xs text-foreground/40 mt-1">
                        Signed in {formatDate(session.issued_at)}
                      </div>
                    </div>
                  </div>
                  {!session.is_current && (
                    <button
                      onClick={() => revokeSessionMutation.mutate(session.id)}
                      disabled={revokeSessionMutation.isPending}
                      className="p-2 hover:bg-red-500/20 text-foreground/40 hover:text-red-400 rounded-lg transition-colors"
                      title="Revoke session"
                    >
                      <Trash2 size={18} />
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Setup Modal */}
      <Modal open={setupModal} onClose={() => {
        setSetupModal(false)
        setSetupData(null)
        setSelectedMethod(null)
        setVerifyCode('')
      }} title="Setup Two-Factor Authentication">
        {!setupData ? (
          <div className="space-y-4">
            <p className="text-foreground/60 mb-4">Choose your preferred 2FA method:</p>

            {/* TOTP Method */}
            <button
              onClick={() => handleSetupMethod('totp')}
              disabled={setupMutation.isPending}
              className="w-full p-4 bg-foreground/5 hover:bg-foreground/10 rounded-xl border border-foreground/10 hover:border-blue-500/30 transition-all flex items-center gap-4 text-left"
            >
              <div className="p-3 bg-blue-500/20 rounded-xl">
                <KeyRound className="w-6 h-6 text-blue-400" />
              </div>
              <div className="flex-1">
                <h3 className="font-medium text-foreground">Authenticator App</h3>
                <p className="text-sm text-foreground/60">Use an app like Google Authenticator or Authy</p>
              </div>
              <ChevronRight className="w-5 h-5 text-foreground/40" />
            </button>

            {/* Email Method */}
            <button
              onClick={() => handleSetupMethod('email')}
              disabled={setupMutation.isPending}
              className="w-full p-4 bg-foreground/5 hover:bg-foreground/10 rounded-xl border border-foreground/10 hover:border-blue-500/30 transition-all flex items-center gap-4 text-left"
            >
              <div className="p-3 bg-purple-500/20 rounded-xl">
                <Mail className="w-6 h-6 text-purple-400" />
              </div>
              <div className="flex-1">
                <h3 className="font-medium text-foreground">Email</h3>
                <p className="text-sm text-foreground/60">Receive codes via email</p>
              </div>
              <ChevronRight className="w-5 h-5 text-foreground/40" />
            </button>

            {/* SMS Method */}
            <button
              onClick={() => handleSetupMethod('sms')}
              disabled={setupMutation.isPending}
              className="w-full p-4 bg-foreground/5 hover:bg-foreground/10 rounded-xl border border-foreground/10 hover:border-blue-500/30 transition-all flex items-center gap-4 text-left"
            >
              <div className="p-3 bg-green-500/20 rounded-xl">
                <Smartphone className="w-6 h-6 text-green-400" />
              </div>
              <div className="flex-1">
                <h3 className="font-medium text-foreground">SMS</h3>
                <p className="text-sm text-foreground/60">Receive codes via text message</p>
              </div>
              <ChevronRight className="w-5 h-5 text-foreground/40" />
            </button>
          </div>
        ) : (
          <div className="space-y-6">
            {/* TOTP Setup */}
            {setupData.method === 'totp' && setupData.provisioning_uri && (
              <>
                <div className="text-center">
                  <p className="text-foreground/60 mb-4">Scan this QR code with your authenticator app:</p>
                  <div className="inline-block p-4 bg-white rounded-xl">
                    <img
                      src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(setupData.provisioning_uri)}`}
                      alt="QR Code"
                      className="w-48 h-48"
                    />
                  </div>
                </div>

                <div className="text-center">
                  <p className="text-foreground/60 text-sm mb-2">Or enter this code manually:</p>
                  <div className="flex items-center justify-center gap-2">
                    <code className="px-3 py-2 bg-foreground/10 rounded-lg font-mono text-foreground">
                      {setupData.secret}
                    </code>
                    <button
                      onClick={() => copyToClipboard(setupData.secret || '')}
                      className="p-2 hover:bg-foreground/10 rounded-lg transition-colors"
                    >
                      <Copy size={18} className="text-foreground/60" />
                    </button>
                  </div>
                </div>
              </>
            )}

            {/* Email/SMS Setup */}
            {(setupData.method === 'email' || setupData.method === 'sms') && (
              <div className="text-center">
                <p className="text-foreground/60">{setupData.message}</p>
              </div>
            )}

            {/* Verification input */}
            <div>
              <label className="block text-sm font-medium text-foreground/80 mb-2">
                Enter verification code
              </label>
              <input
                type="text"
                value={verifyCode}
                onChange={(e) => setVerifyCode(e.target.value)}
                className="w-full px-4 py-3 bg-foreground/10 border border-foreground/20 rounded-xl text-foreground text-center text-xl tracking-widest focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="000000"
                maxLength={6}
              />
            </div>

            <div className="flex gap-3">
              <button
                onClick={() => {
                  setSetupData(null)
                  setSelectedMethod(null)
                  setVerifyCode('')
                }}
                className="flex-1 py-3 px-4 bg-foreground/10 hover:bg-foreground/20 text-foreground font-medium rounded-xl transition-colors"
              >
                Back
              </button>
              <button
                onClick={() => verifyMutation.mutate(verifyCode)}
                disabled={verifyMutation.isPending || !verifyCode.trim()}
                className="flex-1 py-3 px-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 disabled:from-foreground/20 disabled:to-foreground/25 text-white font-semibold rounded-xl transition-all"
              >
                {verifyMutation.isPending ? 'Verifying...' : 'Verify & Enable'}
              </button>
            </div>
          </div>
        )}
      </Modal>

      {/* Disable Modal */}
      <Modal open={disableModal} onClose={() => {
        setDisableModal(false)
        setDisableCode('')
      }} title="Disable Two-Factor Authentication">
        <div className="space-y-4">
          <div className="p-4 bg-yellow-500/10 border border-yellow-500/20 rounded-xl flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-yellow-400 flex-shrink-0 mt-0.5" />
            <p className="text-sm text-yellow-400">
              Disabling 2FA will make your account less secure. Enter your current 2FA code to confirm.
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground/80 mb-2">
              Current 2FA code or backup code
            </label>
            <input
              type="text"
              value={disableCode}
              onChange={(e) => setDisableCode(e.target.value.toUpperCase())}
              className="w-full px-4 py-3 bg-foreground/10 border border-foreground/20 rounded-xl text-foreground text-center text-xl tracking-widest focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="000000"
              maxLength={9}
            />
          </div>

          <div className="flex gap-3">
            <button
              onClick={() => {
                setDisableModal(false)
                setDisableCode('')
              }}
              className="flex-1 py-3 px-4 bg-foreground/10 hover:bg-foreground/20 text-foreground font-medium rounded-xl transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={() => disableMutation.mutate(disableCode)}
              disabled={disableMutation.isPending || !disableCode.trim()}
              className="flex-1 py-3 px-4 bg-red-500 hover:bg-red-600 disabled:bg-foreground/20 text-foreground font-semibold rounded-xl transition-colors"
            >
              {disableMutation.isPending ? 'Disabling...' : 'Disable 2FA'}
            </button>
          </div>
        </div>
      </Modal>

      {/* Backup Codes Modal */}
      <Modal open={backupCodesModal} onClose={() => {
        if (backupCodes.length > 0) {
          setBackupCodesModal(false)
          setBackupCodes([])
        }
      }} title={backupCodes.length > 0 ? "Your Backup Codes" : "Regenerate Backup Codes"}>
        {backupCodes.length > 0 ? (
          <div className="space-y-4">
            <div className="p-4 bg-yellow-500/10 border border-yellow-500/20 rounded-xl flex items-start gap-3">
              <AlertTriangle className="w-5 h-5 text-yellow-400 flex-shrink-0 mt-0.5" />
              <div>
                <p className="text-sm text-yellow-400 font-medium">Save these codes in a safe place</p>
                <p className="text-sm text-yellow-400/80">
                  Each code can only be used once. Store them securely - you won't be able to see them again.
                </p>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {backupCodes.map((code, i) => (
                <div key={i} className="px-3 py-2 bg-foreground/10 rounded-lg font-mono text-foreground text-center">
                  {code}
                </div>
              ))}
            </div>

            <button
              onClick={() => copyToClipboard(backupCodes.join('\n'))}
              className="w-full py-3 px-4 bg-foreground/10 hover:bg-foreground/20 text-foreground font-medium rounded-xl transition-colors flex items-center justify-center gap-2"
            >
              <Copy size={18} />
              Copy All Codes
            </button>

            <button
              onClick={() => {
                setBackupCodesModal(false)
                setBackupCodes([])
              }}
              className="w-full py-3 px-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white font-semibold rounded-xl transition-all"
            >
              I've Saved These Codes
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            <p className="text-foreground/60">
              Enter your current 2FA code to generate new backup codes. This will invalidate all existing backup codes.
            </p>

            <div>
              <label className="block text-sm font-medium text-foreground/80 mb-2">
                Current 2FA code
              </label>
              <input
                type="text"
                value={regenerateCode}
                onChange={(e) => setRegenerateCode(e.target.value)}
                className="w-full px-4 py-3 bg-foreground/10 border border-foreground/20 rounded-xl text-foreground text-center text-xl tracking-widest focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="000000"
                maxLength={6}
              />
            </div>

            <div className="flex gap-3">
              <button
                onClick={() => {
                  setBackupCodesModal(false)
                  setRegenerateCode('')
                }}
                className="flex-1 py-3 px-4 bg-foreground/10 hover:bg-foreground/20 text-foreground font-medium rounded-xl transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() => regenerateMutation.mutate(regenerateCode)}
                disabled={regenerateMutation.isPending || !regenerateCode.trim()}
                className="flex-1 py-3 px-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 disabled:from-foreground/20 disabled:to-foreground/25 text-white font-semibold rounded-xl transition-all"
              >
                {regenerateMutation.isPending ? 'Generating...' : 'Generate New Codes'}
              </button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
