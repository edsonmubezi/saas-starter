import { useQuery } from '@tanstack/react-query'
import { getUserPasswordResetHistory, getOrgUserPasswordResetHistory, downloadPasswordResetForm, downloadOrgPasswordResetForm, type PasswordResetHistoryItem } from '../utils/users'
import Modal from '../ui/Modal'
import { Clock, User, Mail, FileText, Download, CheckCircle, XCircle } from 'lucide-react'

type Props = {
  userId: string
  userName: string
  isOpen: boolean
  onClose: () => void
  orgLevel?: boolean
}

export default function PasswordResetHistoryModal({ userId, userName, isOpen, onClose, orgLevel }: Props) {
  const { data: history, isLoading, error } = useQuery<PasswordResetHistoryItem[]>({
    queryKey: ['password-reset-history', userId],
    queryFn: () => orgLevel ? getOrgUserPasswordResetHistory(userId) : getUserPasswordResetHistory(userId),
    enabled: isOpen && !!userId,
  })

  const handleDownloadPDF = (resetId: number) => {
    const url = orgLevel ? downloadOrgPasswordResetForm(resetId) : downloadPasswordResetForm(resetId)
    window.open(url, '_blank')
  }

  const getMethodBadge = (method: string) => {
    const colors = {
      email: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
      form: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
      both: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
    }
    const icons = {
      email: <Mail className="h-3 w-3" />,
      form: <FileText className="h-3 w-3" />,
      both: <Mail className="h-3 w-3" />,
    }

    return (
      <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${colors[method as keyof typeof colors] || colors.email}`}>
        {icons[method as keyof typeof icons]}
        {method.charAt(0).toUpperCase() + method.slice(1)}
      </span>
    )
  }

  if (!isOpen) return null

  return (
    <Modal
      open={isOpen}
      onClose={onClose}
      title="Password Reset History"
      size="lg"
    >
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <div className="loader-circle h-8 w-8 border-2 border-blue-600"></div>
        </div>
      )}

      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <p className="text-red-800 dark:text-red-200">
            Failed to load password reset history. Please try again.
          </p>
        </div>
      )}

      {history && (
        <div className="space-y-4">
          {/* User Info */}
          <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
            <p className="text-sm text-gray-600 dark:text-foreground/50">
              Password reset history for: <span className="font-semibold text-gray-900 dark:text-gray-100">{userName}</span>
            </p>
            <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
              Total resets: {history.length}
            </p>
          </div>

          {/* History List */}
          {history.length === 0 ? (
            <div className="text-center py-12">
              <FileText className="h-12 w-12 text-foreground/50 mx-auto mb-4" />
              <p className="text-gray-600 dark:text-foreground/50">No password reset history found</p>
              <p className="text-sm text-gray-500 dark:text-gray-500 mt-1">
                This user has never had their password reset by an administrator
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {history.map((item) => (
                <div
                  key={item.id}
                  className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
                >
                  {/* Header Row */}
                  <div className="flex items-start justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <Clock className="h-4 w-4 text-foreground/50" />
                      <span className="text-sm font-medium">
                        {new Date(item.reset_at).toLocaleString()}
                      </span>
                    </div>
                    {getMethodBadge(item.method)}
                  </div>

                  {/* Details Grid */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
                    <div>
                      <div className="flex items-center gap-2 text-xs text-gray-600 dark:text-foreground/50 mb-1">
                        <User className="h-3 w-3" />
                        Reset By
                      </div>
                      <p className="text-sm font-medium">{item.reset_by_name}</p>
                    </div>

                    {item.form_reference && (
                      <div>
                        <div className="flex items-center gap-2 text-xs text-gray-600 dark:text-foreground/50 mb-1">
                          <FileText className="h-3 w-3" />
                          Form Reference
                        </div>
                        <code className="text-sm font-mono bg-gray-100 dark:bg-gray-900 px-2 py-1 rounded">
                          {item.form_reference}
                        </code>
                      </div>
                    )}
                  </div>

                  {/* Status Indicators */}
                  <div className="flex flex-wrap items-center gap-2 mb-3">
                    {item.email_sent && (
                      <span className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400">
                        <CheckCircle className="h-3 w-3" />
                        Email Sent
                      </span>
                    )}
                    {item.pdf_generated && (
                      <span className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400">
                        <CheckCircle className="h-3 w-3" />
                        PDF Generated
                      </span>
                    )}
                    {item.temporary_password_generated && (
                      <span className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400">
                        <CheckCircle className="h-3 w-3" />
                        Temp Password
                      </span>
                    )}
                  </div>

                  {/* Notes */}
                  {item.notes && (
                    <div className="bg-gray-50 dark:bg-gray-900/50 rounded p-3 mb-3">
                      <p className="text-xs text-gray-600 dark:text-foreground/50 mb-1 font-medium">Admin Notes:</p>
                      <p className="text-sm text-gray-700 dark:text-gray-300">{item.notes}</p>
                    </div>
                  )}

                  {/* Expiry Info */}
                  {item.expiry_time && (
                    <div className="text-xs text-gray-500 dark:text-gray-500">
                      <Clock className="h-3 w-3 inline mr-1" />
                      Expires: {new Date(item.expiry_time).toLocaleString()}
                    </div>
                  )}

                  {/* Download PDF Button */}
                  {item.pdf_generated && (
                    <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
                      <button
                        onClick={() => handleDownloadPDF(item.id)}
                        className="px-3 py-1.5 text-xs font-medium text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded transition-colors flex items-center gap-2"
                      >
                        <Download className="h-3 w-3" />
                        Download PDF Form
                      </button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* Close Button */}
          <div className="flex justify-end pt-4 border-t border-gray-200 dark:border-gray-700">
            <button
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      )}
    </Modal>
  )
}
