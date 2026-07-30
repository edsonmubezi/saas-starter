import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Building2, Mail, ArrowLeft, CheckCircle } from 'lucide-react'
import { request } from '../../utils/common'

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [sent, setSent] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (loading) return
    setError(null)
    setLoading(true)

    try {
      await request('/auth/forgot-password', {
        method: 'POST',
        body: JSON.stringify({ email }),
      })
      setSent(true)
    } catch (err: any) {
      // Always show success to prevent email enumeration
      setSent(true)
    } finally {
      setLoading(false)
    }
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
            {sent ? (
              <div className="text-center">
                <div className="mx-auto w-14 h-14 rounded-full bg-green-500/10 border border-green-500/20 flex items-center justify-center mb-4">
                  <CheckCircle className="w-7 h-7 text-green-400" />
                </div>
                <h2 className="text-xl font-bold text-foreground mb-2">Check Your Email</h2>
                <p className="text-foreground/60 text-sm mb-6">
                  If an account exists with <span className="font-medium text-foreground/80">{email}</span>,
                  we've sent a password reset link. Please check your inbox and spam folder.
                </p>
                <p className="text-foreground/40 text-xs mb-6">
                  The link will expire in 30 minutes.
                </p>
                <Link
                  to="/login"
                  className="inline-flex items-center gap-2 text-sm text-blue-400 hover:text-blue-300 transition-colors"
                >
                  <ArrowLeft size={16} />
                  Back to Sign In
                </Link>
              </div>
            ) : (
              <>
                <div className="text-center mb-8">
                  <div className="mx-auto w-14 h-14 rounded-full bg-blue-500/10 border border-blue-500/20 flex items-center justify-center mb-4">
                    <Mail className="w-7 h-7 text-blue-400" />
                  </div>
                  <h2 className="text-2xl font-bold text-foreground mb-2">Forgot Password?</h2>
                  <p className="text-foreground/60 text-sm">
                    Enter your email address and we'll send you a link to reset your password.
                  </p>
                </div>

                <form className="space-y-6" onSubmit={onSubmit}>
                  <div>
                    <label className="block text-sm font-medium text-foreground/80 mb-2">
                      Email Address
                    </label>
                    <input
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      className="w-full px-4 py-3 bg-foreground/10 border border-foreground/20 rounded-xl text-foreground placeholder-foreground/50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                      placeholder="Enter your email address"
                      required
                      autoFocus
                    />
                  </div>

                  {error && (
                    <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20">
                      <p className="text-sm text-red-400">{error}</p>
                    </div>
                  )}

                  <button
                    type="submit"
                    disabled={loading}
                    className="w-full py-3 px-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 disabled:from-foreground/20 disabled:to-foreground/25 text-white font-semibold rounded-xl transition-all duration-200 transform hover:scale-[1.02] hover:shadow-lg hover:shadow-blue-500/25 disabled:cursor-not-allowed disabled:transform-none"
                  >
                    {loading ? (
                      <div className="flex items-center justify-center gap-2">
                        <div className="w-5 h-5 border-2 border-foreground/30 border-t-white rounded-full animate-spin"></div>
                        Sending...
                      </div>
                    ) : (
                      'Send Reset Link'
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
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
