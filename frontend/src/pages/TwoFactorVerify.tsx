import React, { useState, useEffect, useRef } from 'react'
import { useNavigate, useLocation, Navigate } from 'react-router-dom'
import { useAuth } from '../state/AuthContext'
import { request } from '../utils/common'
import { saveTokens, loadTokens } from '../state/tokenStore'
import toast from 'react-hot-toast'
import { Shield, KeyRound, Mail, Smartphone, ChevronLeft, RefreshCw } from 'lucide-react'

interface LocationState {
  session_token: string
  method: string
  next: string
}

/**
 * Decode JWT payload without verification (for expiry extraction)
 */
function decodeJWT(token: string): any {
  try {
    const base64Url = token.split('.')[1]
    if (!base64Url) return null
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch {
    return null
  }
}

export default function TwoFactorVerifyPage() {
  const location = useLocation()
  const nav = useNavigate()
  const { refresh } = useAuth()

  const state = location.state as LocationState | null

  const [code, setCode] = useState('')
  const [loading, setLoading] = useState(false)
  const [resending, setResending] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [useBackupCode, setUseBackupCode] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const inputRef = useRef<HTMLInputElement>(null)

  // If no session token, redirect to login
  if (!state?.session_token) {
    return <Navigate to="/login" replace />
  }

  const { session_token, method, next } = state

  // Focus input on mount
  useEffect(() => {
    inputRef.current?.focus()
  }, [])

  // Countdown timer for resend
  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [countdown])

  // Clear error when code changes
  useEffect(() => {
    if (formError) setFormError(null)
  }, [code])

  async function handleVerify(e: React.FormEvent) {
    e.preventDefault()
    if (loading || !code.trim()) return

    setFormError(null)
    setLoading(true)

    try {
      const response = await request('/auth/2fa/verify-login', {
        method: 'POST',
        body: JSON.stringify({
          session_token,
          code: code.trim(),
          use_backup: useBackupCode
        })
      })

      // Store tokens
      if (response.data?.access_token && response.data?.refresh_token) {
        // Decode token to get expiry time
        const payload = decodeJWT(response.data.access_token)
        const expiresAt = payload?.exp ? payload.exp * 1000 : Date.now() + 3600000
        saveTokens({
          access: response.data.access_token,
          refresh: response.data.refresh_token,
          expiresAt
        })
      }

      await refresh()
      toast.success('Welcome back!')
      nav(next || '/', { replace: true })
    } catch (e: any) {
      setFormError(e?.message || 'Invalid verification code')
    } finally {
      setLoading(false)
    }
  }

  async function handleResendCode() {
    if (resending || countdown > 0) return

    setResending(true)
    try {
      await request('/auth/2fa/send-login-code', {
        method: 'POST',
        body: JSON.stringify({ session_token })
      })
      toast.success('A new code has been sent')
      setCountdown(60) // 60 second cooldown
    } catch (e: any) {
      toast.error(e?.message || 'Failed to resend code')
    } finally {
      setResending(false)
    }
  }

  function getMethodIcon() {
    switch (method) {
      case 'totp':
        return <KeyRound className="w-8 h-8 text-blue-400" />
      case 'email':
        return <Mail className="w-8 h-8 text-blue-400" />
      case 'sms':
        return <Smartphone className="w-8 h-8 text-blue-400" />
      default:
        return <Shield className="w-8 h-8 text-blue-400" />
    }
  }

  function getMethodText() {
    switch (method) {
      case 'totp':
        return 'Enter the 6-digit code from your authenticator app'
      case 'email':
        return 'Enter the 6-digit code sent to your email'
      case 'sms':
        return 'Enter the 6-digit code sent to your phone'
      default:
        return 'Enter your verification code'
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-100 via-blue-50/50 to-indigo-50 dark:from-surface-primary dark:via-surface-secondary dark:to-surface-secondary relative overflow-hidden flex items-center justify-center px-6">
      {/* Background Pattern */}
      <div className="absolute inset-0 opacity-10">
        <div className="absolute top-0 -left-4 w-72 h-72 bg-blue-500 rounded-full mix-blend-multiply filter blur-xl animate-pulse"></div>
        <div className="absolute top-0 -right-4 w-72 h-72 bg-purple-500 rounded-full mix-blend-multiply filter blur-xl animate-pulse animation-delay-2000"></div>
        <div className="absolute -bottom-8 left-20 w-72 h-72 bg-pink-500 rounded-full mix-blend-multiply filter blur-xl animate-pulse animation-delay-4000"></div>
      </div>

      <div className="relative z-10 w-full max-w-md">
        {/* Back to Login */}
        <button
          onClick={() => nav('/login')}
          className="flex items-center gap-2 text-foreground/60 hover:text-foreground mb-6 transition-colors"
        >
          <ChevronLeft size={20} />
          <span>Back to login</span>
        </button>

        {/* 2FA Card */}
        <div className="bg-foreground/5 backdrop-blur-xl rounded-2xl border border-foreground/10 p-8 shadow-2xl">
          {/* Icon */}
          <div className="flex justify-center mb-6">
            <div className="p-4 bg-blue-500/20 rounded-full ring-1 ring-blue-500/30">
              {getMethodIcon()}
            </div>
          </div>

          {/* Title */}
          <div className="text-center mb-6">
            <h2 className="text-2xl font-bold text-foreground mb-2">
              Two-Factor Authentication
            </h2>
            <p className="text-foreground/60">
              {useBackupCode ? 'Enter one of your backup codes' : getMethodText()}
            </p>
          </div>

          <form onSubmit={handleVerify} className="space-y-6">
            {/* Code Input */}
            <div>
              <label className="block text-sm font-medium text-foreground/80 mb-2">
                {useBackupCode ? 'Backup Code' : 'Verification Code'}
              </label>
              <input
                ref={inputRef}
                type="text"
                value={code}
                onChange={(e) => setCode(e.target.value.toUpperCase())}
                className="w-full px-4 py-3 bg-foreground/10 border border-foreground/20 rounded-xl text-foreground text-center text-2xl tracking-widest placeholder-foreground/50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                placeholder={useBackupCode ? 'XXXX-XXXX' : '000000'}
                maxLength={useBackupCode ? 9 : 6}
                autoComplete="one-time-code"
                required
              />
            </div>

            {/* Error Message */}
            {formError && (
              <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20">
                <p className="text-sm text-red-400 text-center">{formError}</p>
              </div>
            )}

            {/* Verify Button */}
            <button
              type="submit"
              disabled={loading || !code.trim()}
              className="w-full py-3 px-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 disabled:from-foreground/20 disabled:to-foreground/25 text-white font-semibold rounded-xl transition-all duration-200 transform hover:scale-[1.02] hover:shadow-lg hover:shadow-blue-500/25 disabled:cursor-not-allowed disabled:transform-none"
            >
              {loading ? (
                <div className="flex items-center justify-center gap-2">
                  <div className="w-5 h-5 border-2 border-foreground/30 border-t-white rounded-full animate-spin"></div>
                  Verifying...
                </div>
              ) : (
                'Verify'
              )}
            </button>

            {/* Resend Code (for email/SMS) */}
            {(method === 'email' || method === 'sms') && !useBackupCode && (
              <button
                type="button"
                onClick={handleResendCode}
                disabled={resending || countdown > 0}
                className="w-full py-2 text-foreground/60 hover:text-foreground text-sm flex items-center justify-center gap-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <RefreshCw size={16} className={resending ? 'animate-spin' : ''} />
                {countdown > 0
                  ? `Resend code in ${countdown}s`
                  : resending
                  ? 'Sending...'
                  : "Didn't receive a code? Resend"}
              </button>
            )}

            {/* Backup Code Toggle */}
            <div className="text-center pt-2">
              <button
                type="button"
                onClick={() => {
                  setUseBackupCode(!useBackupCode)
                  setCode('')
                  setFormError(null)
                }}
                className="text-sm text-blue-400 hover:text-blue-300 transition-colors"
              >
                {useBackupCode
                  ? 'Use verification code instead'
                  : "Can't access your device? Use a backup code"}
              </button>
            </div>
          </form>
        </div>

        {/* Footer */}
        <div className="mt-6 text-center">
          <p className="text-sm text-foreground/40">
            Secure two-factor authentication by SaaS Starter
          </p>
        </div>
      </div>
    </div>
  )
}
