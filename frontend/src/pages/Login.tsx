import React, { useState, useEffect } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { login } from '../utils/api'
import { resetAuthExpired } from '../utils/common'
import { useAuth } from '../state/AuthContext'
import toast from 'react-hot-toast'
import { Eye, EyeOff, Landmark, TrendingUp, ShieldCheck, Users, AlertTriangle, Lock, UserX } from 'lucide-react'
export const API_BASE = import.meta.env.VITE_API_BASE_URL || ''

type ErrorType = 'default' | 'locked' | 'deactivated' | 'warning'

export default function LoginPage() {
  const [identifier, setIdentifier] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [loading, setLoading] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [errorType, setErrorType] = useState<ErrorType>('default')

  const [sp] = useSearchParams()
  // Prevent open redirect attacks - only allow relative paths starting with /
  const rawNext = sp.get('next') || '/'
  const next = rawNext.startsWith('/') && !rawNext.startsWith('//') ? rawNext : '/'

  const nav = useNavigate()
  const { refresh } = useAuth()

  // Clear any stale "auth expired" flag when landing on login page
  useEffect(() => { resetAuthExpired() }, [])

  // Determine error type from message
  function getErrorType(message: string): ErrorType {
    const lowerMessage = message.toLowerCase()
    if (lowerMessage.includes('locked')) return 'locked'
    if (lowerMessage.includes('deactivated')) return 'deactivated'
    if (lowerMessage.includes('attempt') || lowerMessage.includes('remaining')) return 'warning'
    return 'default'
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (loading) return

    setFormError(null)
    setErrorType('default')
    setLoading(true)

    try {
      const result = await login(identifier, password)

      // Handle 2FA requirement
      if (result.requires_2fa) {
        toast('Two-factor authentication required', { icon: '🔐' })
        nav('/2fa-verify', {
          state: {
            session_token: result.session_token,
            method: result.two_factor_method,
            next: next
          }
        })
        return
      }

      const me = await refresh()
      if (!me) {
        setFormError('Failed to load user profile. Please try again.')
        return
      }
      toast.success('Welcome back!')
      nav(next, { replace: true })
    } catch (e: any) {
      const message = e?.message || 'Login failed'
      setFormError(message)
      setErrorType(getErrorType(message))
    } finally {
      setLoading(false)
    }
  }

  const features = [
    { icon: Landmark, title: 'Multi-Tenant', desc: 'Full data isolation per organization' },
    { icon: TrendingUp, title: 'Analytics', desc: 'Real-time dashboards and insights' },
    { icon: Users, title: 'User Management', desc: 'RBAC with granular permissions' },
    { icon: ShieldCheck, title: 'Secure & Compliant', desc: 'Role-based access controls' },
  ]

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-emerald-50/40 to-teal-50 dark:from-[#0b1120] dark:via-[#0d1a2d] dark:to-[#0b1a1e] relative overflow-hidden">
      {/* Background accents */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-24 -left-24 w-96 h-96 bg-emerald-400/8 rounded-full blur-3xl"></div>
        <div className="absolute top-1/3 -right-32 w-80 h-80 bg-teal-400/8 rounded-full blur-3xl"></div>
        <div className="absolute -bottom-20 left-1/3 w-72 h-72 bg-emerald-500/6 rounded-full blur-3xl"></div>
      </div>

      <div className="relative z-10 min-h-screen grid lg:grid-cols-2">
        {/* Left Side - Branding & Features */}
        <div className="hidden lg:flex flex-col justify-center px-12 xl:px-16">
          <div className="max-w-lg">
            {/* Logo & Title */}
            <div className="mb-12">
              <div className="flex items-center gap-3 mb-8">
                <div className="p-3 bg-emerald-500/15 rounded-2xl ring-1 ring-emerald-500/25">
                  <Landmark className="w-8 h-8 text-emerald-500 dark:text-emerald-400" />
                </div>
                <h1 className="text-3xl font-bold text-slate-800 dark:text-white tracking-tight">
                  SaaS Starter
                </h1>
              </div>
              <h2 className="text-4xl font-bold text-slate-900 dark:text-white mb-4 leading-tight">
                Multitenant
                <span className="block text-3xl text-emerald-600 dark:text-emerald-400">
                  Admin Platform
                </span>
              </h2>
              <p className="text-lg text-slate-600 dark:text-slate-400 leading-relaxed">
                Manage your organization with a secure, modern SaaS platform
                built for teams of any size.
              </p>
            </div>

            {/* Features Grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {features.map((feature, idx) => (
                <div key={idx} className="group">
                  <div className="p-4 bg-white/60 dark:bg-white/5 backdrop-blur-sm rounded-xl border border-slate-200/80 dark:border-white/10 hover:border-emerald-400/40 dark:hover:border-emerald-500/30 transition-all duration-300 hover:shadow-sm">
                    <feature.icon className="w-5 h-5 text-emerald-600 dark:text-emerald-400 mb-3 group-hover:scale-110 transition-transform duration-300" />
                    <h3 className="font-semibold text-slate-800 dark:text-white mb-1 text-sm">{feature.title}</h3>
                    <p className="text-xs text-slate-500 dark:text-slate-400">{feature.desc}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Right Side - Login Form */}
        <div className="flex items-center justify-center px-6 lg:px-8">
          <div className="w-full max-w-md">
            {/* Mobile Logo */}
            <div className="lg:hidden text-center mb-8">
              <div className="inline-flex items-center gap-3 mb-4">
                <div className="p-2.5 bg-emerald-500/15 rounded-xl ring-1 ring-emerald-500/25">
                  <Landmark className="w-6 h-6 text-emerald-500 dark:text-emerald-400" />
                </div>
                <h1 className="text-2xl font-bold text-slate-800 dark:text-white tracking-tight">
                  SaaS Starter
                </h1>
              </div>
            </div>

            {/* Login Card */}
            <div className="bg-white/70 dark:bg-white/5 backdrop-blur-xl rounded-2xl border border-slate-200/80 dark:border-white/10 p-8 shadow-xl shadow-slate-200/50 dark:shadow-none">
              <div className="text-center mb-8">
                <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">Welcome Back</h2>
                <p className="text-slate-500 dark:text-slate-400">
                  Sign in to access your dashboard
                </p>
              </div>

              <form className="space-y-5" onSubmit={onSubmit}>
                {/* Identifier Field (Email or TIN) */}
                <div>
                  <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                    Email or TIN
                  </label>
                  <input
                    type="text"
                    value={identifier}
                    onChange={(e) => setIdentifier(e.target.value)}
                    className="w-full px-4 py-3 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-xl text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500/50 transition-all duration-200"
                    placeholder="Enter your email or TIN"
                    required
                  />
                </div>

                {/* Password Field */}
                <div>
                  <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                    Password
                  </label>
                  <div className="relative">
                    <input
                      type={showPassword ? 'text' : 'password'}
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      className="w-full px-4 py-3 pr-12 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-xl text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500/50 transition-all duration-200"
                      placeholder="Enter your password"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-3 top-1/2 transform -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors duration-200"
                    >
                      {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                    </button>
                  </div>
                </div>

                {/* Forgot Password */}
                <div className="flex justify-end -mt-1">
                  <Link
                    to="/forgot-password"
                    className="text-sm text-emerald-600 dark:text-emerald-400 hover:text-emerald-700 dark:hover:text-emerald-300 transition-colors font-medium"
                  >
                    Forgot Password?
                  </Link>
                </div>

                {/* Error Message */}
                {formError && (
                  <div className={`p-4 rounded-xl flex items-start gap-3 ${
                    errorType === 'locked'
                      ? 'bg-orange-50 dark:bg-orange-500/10 border border-orange-200 dark:border-orange-500/20'
                      : errorType === 'deactivated'
                      ? 'bg-slate-50 dark:bg-slate-500/10 border border-slate-200 dark:border-slate-500/20'
                      : errorType === 'warning'
                      ? 'bg-yellow-50 dark:bg-yellow-500/10 border border-yellow-200 dark:border-yellow-500/20'
                      : 'bg-red-50 dark:bg-red-500/10 border border-red-200 dark:border-red-500/20'
                  }`}>
                    {errorType === 'locked' && <Lock className="w-5 h-5 text-orange-500 dark:text-orange-400 flex-shrink-0 mt-0.5" />}
                    {errorType === 'deactivated' && <UserX className="w-5 h-5 text-slate-400 flex-shrink-0 mt-0.5" />}
                    {errorType === 'warning' && <AlertTriangle className="w-5 h-5 text-yellow-500 dark:text-yellow-400 flex-shrink-0 mt-0.5" />}
                    <p className={`text-sm ${
                      errorType === 'locked'
                        ? 'text-orange-600 dark:text-orange-400'
                        : errorType === 'deactivated'
                        ? 'text-slate-500'
                        : errorType === 'warning'
                        ? 'text-yellow-600 dark:text-yellow-400'
                        : 'text-red-600 dark:text-red-400'
                    }`}>{formError}</p>
                  </div>
                )}

                {/* Sign In Button */}
                <button
                  type="submit"
                  disabled={loading}
                  className="w-full py-3 px-4 bg-emerald-600 hover:bg-emerald-700 disabled:bg-slate-300 dark:disabled:bg-slate-700 text-white font-semibold rounded-xl transition-all duration-200 transform hover:scale-[1.02] hover:shadow-lg hover:shadow-emerald-500/25 disabled:cursor-not-allowed disabled:transform-none disabled:shadow-none"
                >
                  {loading ? (
                    <div className="flex items-center justify-center gap-2">
                      <div className="w-5 h-5 border-2 border-emerald-300 border-t-white rounded-full animate-spin"></div>
                      Signing in...
                    </div>
                  ) : (
                    'Sign In'
                  )}
                </button>

                {/* Additional Info */}
                <div className="text-center pt-2">
                  <p className="text-xs text-slate-400 dark:text-slate-500">
                    Secure login powered by SaaS Starter
                  </p>
                </div>
              </form>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
