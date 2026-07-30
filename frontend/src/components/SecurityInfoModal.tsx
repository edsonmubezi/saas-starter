import { useQuery } from '@tanstack/react-query'
import { getUserSecurityInfo, type UserSecurityInfo } from '../utils/users'
import Modal from '../ui/Modal'
import { Shield, Lock, Unlock, Eye, Clock, MapPin, Smartphone, AlertTriangle, CheckCircle } from 'lucide-react'

type Props = {
  userId: string
  isOpen: boolean
  onClose: () => void
}

export default function SecurityInfoModal({ userId, isOpen, onClose }: Props) {
  const { data: securityInfo, isLoading, error } = useQuery<UserSecurityInfo>({
    queryKey: ['user-security', userId],
    queryFn: () => getUserSecurityInfo(userId),
    enabled: isOpen && !!userId,
  })

  if (!isOpen) return null

  return (
    <Modal
      open={isOpen}
      onClose={onClose}
      title="Security Information"
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
            Failed to load security information. Please try again.
          </p>
        </div>
      )}

      {securityInfo && (
        <div className="space-y-6">
          {/* User Header */}
          <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
            <div className="flex items-center gap-3">
              <div className="h-12 w-12 rounded-full bg-blue-600 flex items-center justify-center text-white font-semibold text-lg">
                {securityInfo.full_name.charAt(0).toUpperCase()}
              </div>
              <div>
                <h3 className="font-semibold text-lg">{securityInfo.full_name}</h3>
                <p className="text-sm text-gray-600 dark:text-foreground/50">{securityInfo.email}</p>
              </div>
            </div>
          </div>

          {/* Account Status */}
          <div>
            <h4 className="font-semibold text-sm text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
              <Shield className="h-4 w-4" />
              Account Status
            </h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600 dark:text-foreground/50">Account Status</span>
                  {securityInfo.active_status ? (
                    <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400">
                      <CheckCircle className="h-3 w-3" />
                      Active
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-foreground/50">
                      Disabled
                    </span>
                  )}
                </div>
              </div>

              <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600 dark:text-foreground/50">Lock Status</span>
                  {securityInfo.is_locked ? (
                    <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400">
                      <Lock className="h-3 w-3" />
                      Locked
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400">
                      <Unlock className="h-3 w-3" />
                      Unlocked
                    </span>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Lock Information (if locked) */}
          {securityInfo.is_locked && (
            <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
              <div className="flex items-start gap-3">
                <AlertTriangle className="h-5 w-5 text-red-600 dark:text-red-400 mt-0.5" />
                <div className="flex-1">
                  <h4 className="font-semibold text-red-800 dark:text-red-200 mb-2">Account Locked</h4>
                  {securityInfo.lock_reason && (
                    <p className="text-sm text-red-700 dark:text-red-300 mb-2">
                      <span className="font-medium">Reason:</span> {securityInfo.lock_reason}
                    </p>
                  )}
                  {securityInfo.locked_at && (
                    <p className="text-sm text-red-700 dark:text-red-300">
                      <span className="font-medium">Locked at:</span> {new Date(securityInfo.locked_at).toLocaleString()}
                    </p>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* Login Security */}
          <div>
            <h4 className="font-semibold text-sm text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
              <Eye className="h-4 w-4" />
              Login Security
            </h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm text-gray-600 dark:text-foreground/50">Failed Attempts</span>
                  <span className={`font-semibold ${securityInfo.failed_login_attempts > 0 ? 'text-orange-600 dark:text-orange-400' : 'text-green-600 dark:text-green-400'}`}>
                    {securityInfo.failed_login_attempts}
                  </span>
                </div>
                {securityInfo.failed_login_attempts > 0 && (
                  <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
                    Account locks after 5 failed attempts
                  </p>
                )}
              </div>

              <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600 dark:text-foreground/50">Password Change Required</span>
                  {securityInfo.must_change_password ? (
                    <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400">
                      Yes
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400">
                      No
                    </span>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Last Login Info */}
          <div>
            <h4 className="font-semibold text-sm text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
              <Clock className="h-4 w-4" />
              Last Login Information
            </h4>
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600 dark:text-foreground/50 flex items-center gap-2">
                  <Clock className="h-4 w-4" />
                  Last Login Time
                </span>
                <span className="text-sm font-medium">
                  {securityInfo.last_login_at
                    ? new Date(securityInfo.last_login_at).toLocaleString()
                    : 'Never'
                  }
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600 dark:text-foreground/50 flex items-center gap-2">
                  <MapPin className="h-4 w-4" />
                  Last Login IP
                </span>
                <span className="text-sm font-medium font-mono">
                  {securityInfo.last_login_ip || 'N/A'}
                </span>
              </div>
            </div>
          </div>

          {/* Two-Factor Authentication */}
          <div>
            <h4 className="font-semibold text-sm text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
              <Smartphone className="h-4 w-4" />
              Two-Factor Authentication
            </h4>
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-gray-600 dark:text-foreground/50">2FA Status</span>
                {securityInfo.two_factor_enabled ? (
                  <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400">
                    <CheckCircle className="h-3 w-3" />
                    Enabled
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-foreground/50">
                    Disabled
                  </span>
                )}
              </div>
              {securityInfo.two_factor_enabled && securityInfo.two_factor_method && (
                <p className="text-sm text-gray-600 dark:text-foreground/50">
                  <span className="font-medium">Method:</span> {securityInfo.two_factor_method === 'email' ? 'Email' : 'Authenticator App'}
                </p>
              )}
            </div>
          </div>

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
