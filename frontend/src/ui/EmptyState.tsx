import React from 'react'
import { Link } from 'react-router-dom'
import { Inbox, FileX, Users, Search, AlertCircle, Plus } from 'lucide-react'
import clsx from 'clsx'

type EmptyVariant = 'default' | 'search' | 'error' | 'no-data' | 'no-users'

interface EmptyStateProps {
  /** Pre-defined variant */
  variant?: EmptyVariant
  /** Custom icon (overrides variant icon) */
  icon?: React.ReactNode
  /** Title text */
  title?: string
  /** Description text */
  description?: string
  /** Action button */
  action?: {
    label: string
    to?: string
    onClick?: () => void
    icon?: React.ReactNode
  }
  /** Additional content */
  children?: React.ReactNode
  /** Size variant */
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

const variantConfig: Record<EmptyVariant, { icon: React.ReactNode; title: string; description: string }> = {
  default: {
    icon: <Inbox className="w-12 h-12" />,
    title: 'No data available',
    description: 'There are no items to display at the moment.',
  },
  search: {
    icon: <Search className="w-12 h-12" />,
    title: 'No results found',
    description: 'Try adjusting your search or filter criteria.',
  },
  error: {
    icon: <AlertCircle className="w-12 h-12" />,
    title: 'Something went wrong',
    description: 'We encountered an error loading this data. Please try again.',
  },
  'no-data': {
    icon: <FileX className="w-12 h-12" />,
    title: 'No records yet',
    description: 'Get started by creating your first record.',
  },
  'no-users': {
    icon: <Users className="w-12 h-12" />,
    title: 'No users found',
    description: 'There are no users matching your criteria.',
  },
}

export default function EmptyState({
  variant = 'default',
  icon,
  title,
  description,
  action,
  children,
  size = 'md',
  className,
}: EmptyStateProps) {
  const config = variantConfig[variant]
  const displayIcon = icon || config.icon
  const displayTitle = title || config.title
  const displayDescription = description || config.description

  const sizeClasses = {
    sm: 'py-8',
    md: 'py-12',
    lg: 'py-16',
  }

  const iconSizes = {
    sm: 'w-10 h-10',
    md: 'w-12 h-12',
    lg: 'w-16 h-16',
  }

  return (
    <div
      className={clsx(
        'flex flex-col items-center justify-center text-center',
        sizeClasses[size],
        className
      )}
    >
      <div className={clsx('text-foreground/20 mb-4', iconSizes[size])}>
        {displayIcon}
      </div>

      <h3 className="text-lg font-semibold text-foreground/80 mb-2">
        {displayTitle}
      </h3>

      <p className="text-sm text-foreground/50 max-w-md mb-6">
        {displayDescription}
      </p>

      {action && (
        action.to ? (
          <Link
            to={action.to}
            className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-500 hover:to-blue-600 text-white font-medium shadow-lg shadow-blue-500/20 transition-all"
          >
            {action.icon || <Plus size={18} />}
            {action.label}
          </Link>
        ) : (
          <button
            onClick={action.onClick}
            className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-500 hover:to-blue-600 text-white font-medium shadow-lg shadow-blue-500/20 transition-all"
          >
            {action.icon || <Plus size={18} />}
            {action.label}
          </button>
        )
      )}

      {children}
    </div>
  )
}

// Loading state component
export function LoadingState({ message = 'Loading...' }: { message?: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-12">
      <div className="w-10 h-10 border-2 border-blue-500/30 border-t-blue-500 rounded-full animate-spin mb-4" />
      <p className="text-sm text-foreground/50">{message}</p>
    </div>
  )
}

// Error state component
export function ErrorState({
  title = 'Error loading data',
  message = 'Something went wrong. Please try again.',
  onRetry,
}: {
  title?: string
  message?: string
  onRetry?: () => void
}) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <AlertCircle className="w-12 h-12 text-red-400/50 mb-4" />
      <h3 className="text-lg font-semibold text-foreground/80 mb-2">{title}</h3>
      <p className="text-sm text-foreground/50 max-w-md mb-6">{message}</p>
      {onRetry && (
        <button
          onClick={onRetry}
          className="px-4 py-2 rounded-lg bg-foreground/10 hover:bg-foreground/20 text-foreground/70 hover:text-foreground transition-all"
        >
          Try Again
        </button>
      )}
    </div>
  )
}
