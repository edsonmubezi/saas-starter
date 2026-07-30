import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { adminResetPassword, orgAdminResetPassword, type AdminPasswordResetRequest, type AdminPasswordResetResponse } from '../utils/users'
import Modal from '../ui/Modal'
import toast from 'react-hot-toast'
import { Mail, FileText, Copy, Download, CheckCircle, AlertCircle } from 'lucide-react'

type Props = {
  userId: string
  userName: string
  isOpen: boolean
  onClose: () => void
  orgLevel?: boolean
}

export default function AdvancedPasswordResetModal({ userId, userName, isOpen, onClose, orgLevel }: Props) {
  const queryClient = useQueryClient()
  const [method, setMethod] = useState<'email' | 'form' | 'both'>('email')
  const [generatePassword, setGeneratePassword] = useState(false)
  const [expiryHours, setExpiryHours] = useState(24)
  const [notes, setNotes] = useState('')
  const [resetResponse, setResetResponse] = useState<AdminPasswordResetResponse | null>(null)
  const [copiedPassword, setCopiedPassword] = useState(false)

  const resetMutation = useMutation({
    mutationFn: (payload: AdminPasswordResetRequest) =>
      orgLevel ? orgAdminResetPassword(userId, payload) : adminResetPassword(userId, payload),
    onSuccess: (data) => {
      setResetResponse(data)
      queryClient.invalidateQueries({ queryKey: ['users'] })
      queryClient.invalidateQueries({ queryKey: ['user-security', userId] })
      toast.success(data.message || 'Password reset initiated successfully')
    },
    onError: (error: any) => {
      toast.error(error?.message || 'Failed to reset password')
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    const payload: AdminPasswordResetRequest = {
      method,
      generate_password: generatePassword,
      notes: notes.trim() || undefined,
      expiry_hours: expiryHours,
    }

    resetMutation.mutate(payload)
  }

  const handleCopyPassword = () => {
    if (resetResponse?.temporary_password) {
      navigator.clipboard.writeText(resetResponse.temporary_password)
      setCopiedPassword(true)
      toast.success('Password copied to clipboard')
      setTimeout(() => setCopiedPassword(false), 2000)
    }
  }

  const handleDownloadPDF = () => {
    if (resetResponse?.pdf_url) {
      window.open(resetResponse.pdf_url, '_blank')
    }
  }

  const handleClose = () => {
    setResetResponse(null)
    setMethod('email')
    setGeneratePassword(false)
    setExpiryHours(24)
    setNotes('')
    setCopiedPassword(false)
    onClose()
  }

  if (!isOpen) return null

  // Show success screen if reset was successful
  if (resetResponse) {
    return (
      <Modal
        open={isOpen}
        onClose={handleClose}
        title="Password Reset Successful"
        size="md"
      >
        <div className="space-y-6">
          {/* Success Message */}
          <div className="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <CheckCircle className="h-5 w-5 text-green-600 dark:text-green-400 mt-0.5" />
              <div>
                <h4 className="font-semibold text-green-800 dark:text-green-200">Password Reset Completed</h4>
                <p className="text-sm text-green-700 dark:text-green-300 mt-1">
                  Password reset for <span className="font-medium">{userName}</span> has been initiated successfully.
                </p>
              </div>
            </div>
          </div>

          {/* Form Reference */}
          {resetResponse.form_reference && (
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Form Reference Number
              </label>
              <div className="flex items-center gap-2">
                <code className="flex-1 px-3 py-2 bg-gray-100 dark:bg-gray-800 rounded text-sm font-mono">
                  {resetResponse.form_reference}
                </code>
              </div>
            </div>
          )}

          {/* Temporary Password (if generated) */}
          {resetResponse.temporary_password && (
            <div className="border border-orange-200 dark:border-orange-800 bg-orange-50 dark:bg-orange-900/20 rounded-lg p-4">
              <div className="flex items-start gap-3 mb-3">
                <AlertCircle className="h-5 w-5 text-orange-600 dark:text-orange-400 mt-0.5" />
                <div className="flex-1">
                  <h4 className="font-semibold text-orange-800 dark:text-orange-200">Temporary Password Generated</h4>
                  <p className="text-sm text-orange-700 dark:text-orange-300 mt-1">
                    Make sure to securely share this password with the user. It will not be shown again.
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-2 mt-3">
                <code className="flex-1 px-3 py-2 bg-white dark:bg-gray-900 border border-orange-200 dark:border-orange-800 rounded text-sm font-mono select-all">
                  {resetResponse.temporary_password}
                </code>
                <button
                  type="button"
                  onClick={handleCopyPassword}
                  className="px-3 py-2 text-sm font-medium text-white bg-orange-600 hover:bg-orange-700 rounded-lg transition-colors flex items-center gap-2"
                >
                  {copiedPassword ? (
                    <>
                      <CheckCircle className="h-4 w-4" />
                      Copied
                    </>
                  ) : (
                    <>
                      <Copy className="h-4 w-4" />
                      Copy
                    </>
                  )}
                </button>
              </div>
              <p className="text-xs text-orange-600 dark:text-orange-400 mt-2">
                User must change this password on first login.
              </p>
            </div>
          )}

          {/* Status Grid */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {resetResponse.email_sent && (
              <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div className="flex items-center gap-2 mb-1">
                  <Mail className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                  <span className="text-sm font-medium">Email Sent</span>
                </div>
                <p className="text-xs text-gray-600 dark:text-foreground/50">Reset link sent to user's email</p>
              </div>
            )}

            {resetResponse.pdf_generated && (
              <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div className="flex items-center gap-2 mb-1">
                  <FileText className="h-4 w-4 text-green-600 dark:text-green-400" />
                  <span className="text-sm font-medium">PDF Generated</span>
                </div>
                <p className="text-xs text-gray-600 dark:text-foreground/50">Reset form is ready for download</p>
              </div>
            )}
          </div>

          {/* Download PDF Button */}
          {resetResponse.pdf_generated && resetResponse.pdf_url && (
            <button
              type="button"
              onClick={handleDownloadPDF}
              className="w-full px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors flex items-center justify-center gap-2"
            >
              <Download className="h-4 w-4" />
              Download Password Reset Form (PDF)
            </button>
          )}

          {/* Close Button */}
          <div className="flex justify-end pt-4 border-t border-gray-200 dark:border-gray-700">
            <button
              onClick={handleClose}
              className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      </Modal>
    )
  }

  // Show form
  return (
    <Modal
      open={isOpen}
      onClose={handleClose}
      title="Advanced Password Reset"
      size="md"
    >
      <form onSubmit={handleSubmit} className="space-y-6">
        {/* User Info */}
        <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
          <p className="text-sm text-gray-600 dark:text-foreground/50">
            Resetting password for: <span className="font-semibold text-gray-900 dark:text-gray-100">{userName}</span>
          </p>
        </div>

        {/* Reset Method */}
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Reset Method <span className="text-red-500">*</span>
          </label>
          <div className="space-y-2">
            <label className="flex items-start gap-3 p-3 border border-gray-200 dark:border-gray-700 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
              <input
                type="radio"
                name="method"
                value="email"
                checked={method === 'email'}
                onChange={(e) => setMethod(e.target.value as any)}
                className="mt-0.5"
              />
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <Mail className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                  <span className="font-medium text-sm">Email Reset Link</span>
                </div>
                <p className="text-xs text-gray-600 dark:text-foreground/50">
                  Send password reset link via email
                </p>
              </div>
            </label>

            <label className="flex items-start gap-3 p-3 border border-gray-200 dark:border-gray-700 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
              <input
                type="radio"
                name="method"
                value="form"
                checked={method === 'form'}
                onChange={(e) => setMethod(e.target.value as any)}
                className="mt-0.5"
              />
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <FileText className="h-4 w-4 text-green-600 dark:text-green-400" />
                  <span className="font-medium text-sm">Generate PDF Form</span>
                </div>
                <p className="text-xs text-gray-600 dark:text-foreground/50">
                  Create official password reset form for manual processing
                </p>
              </div>
            </label>

            <label className="flex items-start gap-3 p-3 border border-gray-200 dark:border-gray-700 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
              <input
                type="radio"
                name="method"
                value="both"
                checked={method === 'both'}
                onChange={(e) => setMethod(e.target.value as any)}
                className="mt-0.5"
              />
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <Mail className="h-4 w-4 text-purple-600 dark:text-purple-400" />
                  <span className="font-medium text-sm">Both Email & Form</span>
                </div>
                <p className="text-xs text-gray-600 dark:text-foreground/50">
                  Send email and generate PDF form
                </p>
              </div>
            </label>
          </div>
        </div>

        {/* Generate Temporary Password */}
        <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
          <label className="flex items-start gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={generatePassword}
              onChange={(e) => setGeneratePassword(e.target.checked)}
              className="mt-0.5"
            />
            <div className="flex-1">
              <span className="font-medium text-sm">Generate Temporary Password</span>
              <p className="text-xs text-gray-600 dark:text-foreground/50 mt-1">
                Create a secure temporary password and update it immediately. User will be forced to change it on first login.
              </p>
            </div>
          </label>
        </div>

        {/* Expiry Hours */}
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Reset Link Expiry (hours)
          </label>
          <input
            type="number"
            min="1"
            max="168"
            value={expiryHours}
            onChange={(e) => setExpiryHours(parseInt(e.target.value) || 24)}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 focus:ring-2 focus:ring-blue-500"
          />
          <p className="text-xs text-gray-600 dark:text-foreground/50 mt-1">
            Default: 24 hours. Maximum: 168 hours (7 days)
          </p>
        </div>

        {/* Admin Notes */}
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Admin Notes (Optional)
          </label>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            rows={3}
            placeholder="Add any notes about why this reset was needed..."
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 focus:ring-2 focus:ring-blue-500 resize-none"
          />
        </div>

        {/* Action Buttons */}
        <div className="flex items-center justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
          <button
            type="button"
            onClick={handleClose}
            className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
            disabled={resetMutation.isPending}
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={resetMutation.isPending}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 disabled:cursor-not-allowed rounded-lg transition-colors flex items-center gap-2"
          >
            {resetMutation.isPending ? (
              <>
                <div className="loader-circle h-4 w-4 border-2 border-white"></div>
                Processing...
              </>
            ) : (
              'Reset Password'
            )}
          </button>
        </div>
      </form>
    </Modal>
  )
}
