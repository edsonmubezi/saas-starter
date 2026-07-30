import React from 'react'
import { CheckCircle, AlertCircle, MinusCircle, Check, X, XCircle } from 'lucide-react'
import clsx from 'clsx'

type StatusType =
  | 'user'       // active, inactive, disabled
  | 'generic'    // success, warning, error, info, neutral

interface StatusBadgeProps {
  /** The status value */
  status?: string | boolean | null
  /** Type of status for context-aware styling */
  type?: StatusType
  /** Override the display label */
  label?: string
  /** Show icon */
  showIcon?: boolean
  /** Size variant */
  size?: 'sm' | 'md'
  className?: string
}

interface StatusConfig {
  label: string
  className: string
  icon?: React.ReactNode
}

// Status configurations by type
const statusConfigs: Record<string, StatusConfig> = {
  // Generic statuses
  active: {
    label: 'Active',
    className: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
    icon: <CheckCircle className="w-3 h-3" />,
  },
  inactive: {
    label: 'Inactive',
    className: 'bg-gray-500/20 text-foreground/50 border-foreground/20',
    icon: <MinusCircle className="w-3 h-3" />,
  },
  disabled: {
    label: 'Disabled',
    className: 'bg-red-500/20 text-red-400 border-red-500/30',
    icon: <XCircle className="w-3 h-3" />,
  },
  pending: {
    label: 'Pending',
    className: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
    icon: <Clock className="w-3 h-3" />,
  },
  approved: {
    label: 'Approved',
    className: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
    icon: <Check className="w-3 h-3" />,
  },
  rejected: {
    label: 'Rejected',
    className: 'bg-red-500/20 text-red-400 border-red-500/30',
    icon: <X className="w-3 h-3" />,
  },
  completed: {
    label: 'Completed',
    className: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
    icon: <CheckCircle className="w-3 h-3" />,
  },

  // Generic
  success: {
    label: 'Success',
    className: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
    icon: <CheckCircle className="w-3 h-3" />,
  },
  warning: {
    label: 'Warning',
    className: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
    icon: <AlertCircle className="w-3 h-3" />,
  },
  error: {
    label: 'Error',
    className: 'bg-red-500/20 text-red-400 border-red-500/30',
    icon: <XCircle className="w-3 h-3" />,
  },
  info: {
    label: 'Info',
    className: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
    icon: <AlertCircle className="w-3 h-3" />,
  },
  neutral: {
    label: 'Unknown',
    className: 'bg-foreground/10 text-foreground/60 border-foreground/20',
    icon: <MinusCircle className="w-3 h-3" />,
  },
}

function normalizeStatus(status?: string | boolean | null): string {
  if (status === null || status === undefined) return 'neutral'
  if (typeof status === 'boolean') return status ? 'active' : 'inactive'

  const s = status.toString().toLowerCase().trim().replace(/\s+/g, '_')

  // Map common variations
  if (s.includes('active') && !s.includes('inactive')) return 'active'
  if (s.includes('inactive')) return 'inactive'
  if (s.includes('disabled')) return 'disabled'
  if (s.includes('pending')) return 'pending'
  if (s.includes('approved')) return 'approved'
  if (s.includes('rejected')) return 'rejected'
  if (s.includes('completed')) return 'completed'

  return s
}

export default function StatusBadge({
  status,
  type = 'generic',
  label,
  showIcon = true,
  size = 'sm',
  className,
}: StatusBadgeProps) {
  const normalizedStatus = normalizeStatus(status)
  const config = statusConfigs[normalizedStatus] || statusConfigs.neutral

  const sizeClasses = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-2.5 py-1 text-sm',
  }

  return (
    <span
      className={clsx(
        'inline-flex items-center gap-1 rounded-full border font-medium',
        config.className,
        sizeClasses[size],
        className
      )}
    >
      {showIcon && config.icon}
      {label || config.label}
    </span>
  )
}

// Export a simple active/inactive badge for backward compatibility
export function ActiveBadge({ value }: { value?: string | boolean }) {
  return <StatusBadge status={value} type="user" />
}
