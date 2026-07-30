import React, { useState, useEffect } from 'react'
import { Lock, Eye, EyeOff, AlertTriangle, User } from 'lucide-react'
import { useAuth } from '../state/AuthContext'
import { request, tryRefresh } from '../utils/common'
import { isTokenExpired } from '../state/tokenStore'
import toast from 'react-hot-toast'

interface LockScreenProps {
  onUnlock: () => void
  onLogout: () => void
}

export default function LockScreen({ onUnlock, onLogout }: LockScreenProps) {
  const { user } = useAuth()
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [attempts, setAttempts] = useState(0)
  const maxAttempts = 3

  // Clear error when password changes
  useEffect(() => {
    if (error) setError(null)
  }, [password])

  async function handleUnlock(e: React.FormEvent) {
    e.preventDefault()
    if (loading || !password.trim()) return

    setError(null)
    setLoading(true)

    try {
      // If access token is expired, refresh it first so verify-password doesn't fail
      if (isTokenExpired()) {
        const refreshed = await tryRefresh()
        if (!refreshed) {
          setError('Your session has expired. Please log in again.')
          toast.error('Session expired')
          setTimeout(() => { onLogout() }, 1500)
          return
        }
      }

      // Verify password with backend
      await request('/auth/verify-password', {
        method: 'POST',
        body: JSON.stringify({ password }),
      })

      // Password verified, unlock the screen
      toast.success('Welcome back!')
      setPassword('')
      setAttempts(0)
      onUnlock()
    } catch (err: any) {
      const status = err?.status ?? err?.response?.status
      // If JWT expired (401/403), auto-logout — user must re-login
      if (status === 401 || status === 403) {
        setError('Your session has expired. Please log in again.')
        toast.error('Session expired')
        setTimeout(() => { onLogout() }, 1500)
        return
      }

      const newAttempts = attempts + 1
      setAttempts(newAttempts)

      if (newAttempts >= maxAttempts) {
        setError('Too many failed attempts. You will be logged out.')
        toast.error('Session expired due to failed unlock attempts')
        setTimeout(() => { onLogout() }, 2000)
      } else {
        setError(`Incorrect password. ${maxAttempts - newAttempts} attempt(s) remaining.`)
      }
    } finally {
      setLoading(false)
    }
  }

  const displayName = user?.fullname || user?.email || 'User'
  const initials = displayName.slice(0, 2).toUpperCase()

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-gradient-to-br from-slate-100 via-blue-50/50 to-indigo-50 dark:from-surface-primary dark:via-surface-secondary dark:to-surface-secondary">
      {/* Background Pattern */}
      <div className="absolute inset-0 opacity-10 dark:opacity-10">
        <div className="absolute top-0 -left-4 w-72 h-72 bg-blue-500 rounded-full mix-blend-multiply filter blur-xl animate-pulse"></div>
        <div className="absolute top-0 -right-4 w-72 h-72 bg-purple-500 rounded-full mix-blend-multiply filter blur-xl animate-pulse animation-delay-2000"></div>
        <div className="absolute -bottom-8 left-20 w-72 h-72 bg-pink-500 rounded-full mix-blend-multiply filter blur-xl animate-pulse animation-delay-4000"></div>
      </div>

      <div className="relative z-10 w-full max-w-md px-6">
        <div className="bg-foreground/5 backdrop-blur-xl rounded-2xl border border-foreground/10 p-8 shadow-2xl">
          {/* Lock Icon */}
          <div className="flex justify-center mb-6">
            <div className="p-4 bg-yellow-500/20 rounded-full ring-1 ring-yellow-500/30">
              <Lock className="w-8 h-8 text-yellow-400" />
            </div>
          </div>

          {/* Title */}
          <div className="text-center mb-6">
            <h2 className="text-2xl font-bold text-foreground mb-2">Session Locked</h2>
            <p className="text-foreground/60 text-sm">
              Your session was locked due to inactivity. Enter your password to continue.
            </p>
          </div>

          {/* User Avatar */}
          <div className="flex items-center justify-center gap-3 mb-6 p-4 bg-foreground/5 rounded-xl">
            <div className="w-12 h-12 rounded-full bg-blue-500/20 flex items-center justify-center text-blue-400 text-lg font-semibold">
              {initials}
            </div>
            <div className="text-left">
              <div className="text-foreground font-medium">{displayName}</div>
              <div className="text-foreground/60 text-sm">{user?.email}</div>
            </div>
          </div>

          <form onSubmit={handleUnlock} className="space-y-4">
            {/* Password Field */}
            <div>
              <label className="block text-sm font-medium text-foreground/80 mb-2">
                Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full px-4 py-3 pr-12 bg-foreground/10 border border-foreground/20 rounded-xl text-foreground placeholder-foreground/50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                  placeholder="Enter your password"
                  autoFocus
                  disabled={loading || attempts >= maxAttempts}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-foreground/50 hover:text-foreground/80 transition-colors duration-200"
                  tabIndex={-1}
                >
                  {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                </button>
              </div>
            </div>

            {/* Error Message */}
            {error && (
              <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-start gap-3">
                <AlertTriangle className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
                <p className="text-sm text-red-400">{error}</p>
              </div>
            )}

            {/* Unlock Button */}
            <button
              type="submit"
              disabled={loading || !password.trim() || attempts >= maxAttempts}
              className="w-full py-3 px-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 disabled:from-foreground/20 disabled:to-foreground/25 text-white font-semibold rounded-xl transition-all duration-200 transform hover:scale-[1.02] hover:shadow-lg hover:shadow-blue-500/25 disabled:cursor-not-allowed disabled:transform-none"
            >
              {loading ? (
                <div className="flex items-center justify-center gap-2">
                  <div className="w-5 h-5 border-2 border-foreground/30 border-t-white rounded-full animate-spin"></div>
                  Verifying...
                </div>
              ) : (
                'Unlock'
              )}
            </button>

            {/* Logout Option */}
            <button
              type="button"
              onClick={onLogout}
              disabled={loading}
              className="w-full py-2 px-4 text-foreground/60 hover:text-foreground text-sm transition-colors duration-200"
            >
              Not you? Sign out and switch account
            </button>
          </form>
        </div>

        {/* Footer */}
        <div className="mt-6 text-center">
          <p className="text-sm text-foreground/40">
            Secure session lock by SaaS Starter
          </p>
        </div>
      </div>
    </div>
  )
}
