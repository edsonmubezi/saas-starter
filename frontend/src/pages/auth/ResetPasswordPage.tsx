import { useState, useEffect } from 'react'
import { Link, useSearchParams, useNavigate } from 'react-router-dom'
import { Building2, Lock, ArrowLeft, CheckCircle, XCircle, Eye, EyeOff } from 'lucide-react'
import { request } from '../../utils/common'
import toast from 'react-hot-toast'

type TokenStatus = 'loading' | 'valid' | 'invalid' | 'expired'

const PASSWORD_RULES = [
  { label: 'At least 12 characters', test: (p: string) => p.length >= 12 },
  { label: 'One uppercase letter', test: (p: string) => /[A-Z]/.test(p) },
  { label: 'One lowercase letter', test: (p: string) => /[a-z]/.test(p) },
  { label: 'One number', test: (p: string) => /[0-9]/.test(p) },
  { label: 'One special character', test: (p: string) => /[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?]/.test(p) },
]

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const token = searchParams.get('token') || ''

  const [tokenStatus, setTokenStatus] = useState<TokenStatus>('loading')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Verify token on mount
  useEffect(() => {
    if (!token) {
      setTokenStatus('invalid')
      return
    }

    request(`/auth/verify-reset-token/${encodeURIComponent(token)}`, { method: 'GET' })
      .then(() => setTokenStatus('valid'))
      .catch((err: any) => {
        const msg = err?.message?.toLowerCase() || ''
        if (msg.includes('expired')) setTokenStatus('expired')
        else setTokenStatus('invalid')
      })
  }, [token])

  const allRulesPass = PASSWORD_RULES.every(r => r.test(password))
  const passwordsMatch = password === confirmPassword && confirmPassword.length > 0

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (loading) return

    if (!allRulesPass) {
      setError('Password does not meet all requirements')
      return
    }
    if (!passwordsMatch) {
      setError('Passwords do not match')
      return
    }

    setError(null)
    setLoading(true)

    try {
      await request('/auth/reset-password', {
        method: 'POST',
        body: JSON.stringify({ token, new_password: password }),
      })
      setSuccess(true)
      toast.success('Password reset successfully')
      setTimeout(() => navigate('/login'), 3000)
    } catch (err: any) {
      setError(err?.message || 'Failed to reset password. The link may have expired.')
    } finally {
      setLoading(false)
    }
  }

  const renderContent = () => {
    // Success state
    if (success) {
      return (
        <div className="text-center">
          <div className="mx-auto w-14 h-14 rounded-full bg-green-500/10 border border-green-500/20 flex items-center justify-center mb-4">
            <CheckCircle className="w-7 h-7 text-green-400" />
          </div>
          <h2 className="text-xl font-bold text-foreground mb-2">Password Reset Successful</h2>
          <p className="text-foreground/60 text-sm mb-6">
            Your password has been reset successfully. You'll be redirected to the login page shortly.
          </p>
          <Link
            to="/login"
            className="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 text-white font-semibold rounded-xl hover:from-blue-700 hover:to-purple-700 transition-all"
          >
            Sign In Now
          </Link>
        </div>
      )
    }

    // Loading token verification
    if (tokenStatus === 'loading') {
      return (
        <div className="text-center py-8">
          <div className="mx-auto w-10 h-10 border-2 border-blue-500 border-t-transparent rounded-full animate-spin mb-4"></div>
          <p className="text-foreground/60 text-sm">Verifying your reset link...</p>
        </div>
      )
    }

    // Invalid or expired token
    if (tokenStatus === 'invalid' || tokenStatus === 'expired') {
      return (
        <div className="text-center">
          <div className="mx-auto w-14 h-14 rounded-full bg-red-500/10 border border-red-500/20 flex items-center justify-center mb-4">
            <XCircle className="w-7 h-7 text-red-400" />
          </div>
          <h2 className="text-xl font-bold text-foreground mb-2">
            {tokenStatus === 'expired' ? 'Link Expired' : 'Invalid Link'}
          </h2>
          <p className="text-foreground/60 text-sm mb-6">
            {tokenStatus === 'expired'
              ? 'This password reset link has expired. Please request a new one.'
              : 'This password reset link is invalid or has already been used.'}
          </p>
          <div className="flex flex-col items-center gap-3">
            <Link
              to="/forgot-password"
              className="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 text-white font-semibold rounded-xl hover:from-blue-700 hover:to-purple-700 transition-all"
            >
              Request New Link
            </Link>
            <Link
              to="/login"
              className="inline-flex items-center gap-2 text-sm text-foreground/50 hover:text-foreground/80 transition-colors"
            >
              <ArrowLeft size={16} />
              Back to Sign In
            </Link>
          </div>
        </div>
      )
    }

    // Valid token - show reset form
    return (
      <>
        <div className="text-center mb-8">
          <div className="mx-auto w-14 h-14 rounded-full bg-blue-500/10 border border-blue-500/20 flex items-center justify-center mb-4">
            <Lock className="w-7 h-7 text-blue-400" />
          </div>
          <h2 className="text-2xl font-bold text-foreground mb-2">Reset Your Password</h2>
          <p className="text-foreground/60 text-sm">
            Enter your new password below.
          </p>
        </div>

        <form className="space-y-6" onSubmit={onSubmit}>
          {/* New Password */}
          <div>
            <label className="block text-sm font-medium text-foreground/80 mb-2">
              New Password
            </label>
            <div className="relative">
              <input
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-3 pr-12 bg-foreground/10 border border-foreground/20 rounded-xl text-foreground placeholder-foreground/50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                placeholder="Enter new password"
                required
                autoFocus
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 transform -translate-y-1/2 text-foreground/50 hover:text-foreground/80 transition-colors"
              >
                {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
              </button>
            </div>
          </div>

          {/* Password Strength Requirements */}
          {password.length > 0 && (
            <div className="p-4 rounded-xl bg-foreground/5 border border-foreground/10">
              <p className="text-xs font-medium text-foreground/60 mb-3">Password Requirements:</p>
              <div className="grid grid-cols-1 gap-1.5">
                {PASSWORD_RULES.map((rule, i) => {
                  const passes = rule.test(password)
                  return (
                    <div key={i} className="flex items-center gap-2">
                      {passes ? (
                        <CheckCircle size={14} className="text-green-400 shrink-0" />
                      ) : (
                        <XCircle size={14} className="text-foreground/30 shrink-0" />
                      )}
                      <span className={`text-xs ${passes ? 'text-green-400' : 'text-foreground/40'}`}>
                        {rule.label}
                      </span>
                    </div>
                  )
                })}
              </div>
            </div>
          )}

          {/* Confirm Password */}
          <div>
            <label className="block text-sm font-medium text-foreground/80 mb-2">
              Confirm Password
            </label>
            <div className="relative">
              <input
                type={showConfirm ? 'text' : 'password'}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="w-full px-4 py-3 pr-12 bg-foreground/10 border border-foreground/20 rounded-xl text-foreground placeholder-foreground/50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                placeholder="Confirm new password"
                required
              />
              <button
                type="button"
                onClick={() => setShowConfirm(!showConfirm)}
                className="absolute right-3 top-1/2 transform -translate-y-1/2 text-foreground/50 hover:text-foreground/80 transition-colors"
              >
                {showConfirm ? <EyeOff size={20} /> : <Eye size={20} />}
              </button>
            </div>
            {confirmPassword.length > 0 && !passwordsMatch && (
              <p className="text-xs text-red-400 mt-2">Passwords do not match</p>
            )}
          </div>

          {error && (
            <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20">
              <p className="text-sm text-red-400">{error}</p>
            </div>
          )}

          <button
            type="submit"
            disabled={loading || !allRulesPass || !passwordsMatch}
            className="w-full py-3 px-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 disabled:from-foreground/20 disabled:to-foreground/25 text-white font-semibold rounded-xl transition-all duration-200 transform hover:scale-[1.02] hover:shadow-lg hover:shadow-blue-500/25 disabled:cursor-not-allowed disabled:transform-none"
          >
            {loading ? (
              <div className="flex items-center justify-center gap-2">
                <div className="w-5 h-5 border-2 border-foreground/30 border-t-white rounded-full animate-spin"></div>
                Resetting...
              </div>
            ) : (
              'Reset Password'
            )}
          </button>

          <div className="text-center">
            <Link
              to="/login"
              className="inline-flex items-center gap-2 text-sm text-foreground/50 hover:text-foreground/80 transition-colors"
            >
              <ArrowLeft size={16} />
              Back to Sign In
            </Link>
          </div>
        </form>
      </>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-100 via-blue-50/50 to-indigo-50 dark:from-surface-primary dark:via-surface-secondary dark:to-surface-secondary relative overflow-hidden">
      {/* Background Pattern */}
      <div className="absolute inset-0 opacity-10">
        <div className="absolute top-0 -left-4 w-72 h-72 bg-blue-500 rounded-full mix-blend-multiply filter blur-xl animate-pulse"></div>
        <div className="absolute top-0 -right-4 w-72 h-72 bg-purple-500 rounded-full mix-blend-multiply filter blur-xl animate-pulse animation-delay-2000"></div>
        <div className="absolute -bottom-8 left-20 w-72 h-72 bg-pink-500 rounded-full mix-blend-multiply filter blur-xl animate-pulse animation-delay-4000"></div>
      </div>

      <div className="relative z-10 min-h-screen flex items-center justify-center px-6">
        <div className="w-full max-w-md">
          {/* Logo */}
          <div className="text-center mb-8">
            <div className="inline-flex items-center gap-3 mb-4">
              <div className="p-2 bg-blue-600/20 rounded-lg ring-1 ring-blue-500/30">
                <Building2 className="w-6 h-6 text-blue-400" />
              </div>
              <h1 className="text-2xl font-bold bg-gradient-to-r from-slate-800 to-blue-600 dark:from-white dark:to-blue-300 bg-clip-text text-transparent">
                SaaS Starter
              </h1>
            </div>
          </div>

          {/* Card */}
          <div className="bg-foreground/5 backdrop-blur-xl rounded-2xl border border-foreground/10 p-8 shadow-2xl">
            {renderContent()}
          </div>
        </div>
      </div>
    </div>
  )
}
