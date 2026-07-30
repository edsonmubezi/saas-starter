import React from 'react'
import { Link } from 'react-router-dom'
import { Eye, Pencil, Trash2, Check, X, Download, Send, MoreHorizontal } from 'lucide-react'
import clsx from 'clsx'

type ActionType = 'view' | 'edit' | 'delete' | 'approve' | 'reject' | 'download' | 'send' | 'custom'

interface Action {
  type: ActionType
  /** For navigation actions */
  to?: string
  /** For button actions */
  onClick?: () => void
  /** Show/hide based on permission */
  show?: boolean
  /** Disable the action */
  disabled?: boolean
  /** Custom label (overrides default) */
  label?: string
  /** Custom icon (for type='custom') */
  icon?: React.ReactNode
  /** Loading state */
  loading?: boolean
}

interface ActionButtonsProps {
  actions: Action[]
  /** Size variant */
  size?: 'sm' | 'md'
  /** Layout direction */
  direction?: 'row' | 'column'
  className?: string
}

const actionConfig: Record<ActionType, { icon: React.ReactNode; label: string; className: string }> = {
  view: {
    icon: <Eye size={16} />,
    label: 'View',
    className: 'bg-blue-500/20 hover:bg-blue-500/30 text-blue-400 border-blue-500/30',
  },
  edit: {
    icon: <Pencil size={16} />,
    label: 'Edit',
    className: 'bg-purple-500/20 hover:bg-purple-500/30 text-purple-400 border-purple-500/30',
  },
  delete: {
    icon: <Trash2 size={16} />,
    label: 'Delete',
    className: 'bg-red-500/20 hover:bg-red-500/30 text-red-400 border-red-500/30',
  },
  approve: {
    icon: <Check size={16} />,
    label: 'Approve',
    className: 'bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-400 border-emerald-500/30',
  },
  reject: {
    icon: <X size={16} />,
    label: 'Reject',
    className: 'bg-red-500/20 hover:bg-red-500/30 text-red-400 border-red-500/30',
  },
  download: {
    icon: <Download size={16} />,
    label: 'Download',
    className: 'bg-cyan-500/20 hover:bg-cyan-500/30 text-cyan-400 border-cyan-500/30',
  },
  send: {
    icon: <Send size={16} />,
    label: 'Send',
    className: 'bg-blue-500/20 hover:bg-blue-500/30 text-blue-400 border-blue-500/30',
  },
  custom: {
    icon: <MoreHorizontal size={16} />,
    label: 'Action',
    className: 'bg-foreground/10 hover:bg-foreground/20 text-foreground/70 border-foreground/20',
  },
}

export default function ActionButtons({
  actions,
  size = 'sm',
  direction = 'row',
  className,
}: ActionButtonsProps) {
  const visibleActions = actions.filter(a => a.show !== false)

  if (visibleActions.length === 0) return null

  const sizeClasses = {
    sm: 'p-2',
    md: 'p-2.5',
  }

  return (
    <div
      className={clsx(
        'flex gap-2',
        direction === 'column' ? 'flex-col' : 'flex-row',
        className
      )}
    >
      {visibleActions.map((action, idx) => {
        const config = actionConfig[action.type]
        const icon = action.icon || config.icon
        const label = action.label || config.label

        const buttonClass = clsx(
          'rounded-lg border transition-all flex items-center justify-center',
          sizeClasses[size],
          config.className,
          action.disabled && 'opacity-50 cursor-not-allowed',
          action.loading && 'animate-pulse'
        )

        if (action.to && !action.disabled) {
          return (
            <Link
              key={idx}
              to={action.to}
              className={buttonClass}
              title={label}
            >
              {icon}
            </Link>
          )
        }

        return (
          <button
            key={idx}
            onClick={action.onClick}
            disabled={action.disabled || action.loading}
            className={buttonClass}
            title={label}
          >
            {action.loading ? (
              <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
            ) : (
              icon
            )}
          </button>
        )
      })}
    </div>
  )
}

// Individual action button exports for flexibility
export function ViewButton({ to, onClick, disabled }: { to?: string; onClick?: () => void; disabled?: boolean }) {
  const className = 'p-2 rounded-lg border bg-blue-500/20 hover:bg-blue-500/30 text-blue-400 border-blue-500/30 transition-all'

  if (to) {
    return (
      <Link to={to} className={className} title="View">
        <Eye size={16} />
      </Link>
    )
  }

  return (
    <button onClick={onClick} disabled={disabled} className={className} title="View">
      <Eye size={16} />
    </button>
  )
}

export function EditButton({ to, onClick, disabled }: { to?: string; onClick?: () => void; disabled?: boolean }) {
  const className = 'p-2 rounded-lg border bg-purple-500/20 hover:bg-purple-500/30 text-purple-400 border-purple-500/30 transition-all'

  if (to) {
    return (
      <Link to={to} className={className} title="Edit">
        <Pencil size={16} />
      </Link>
    )
  }

  return (
    <button onClick={onClick} disabled={disabled} className={className} title="Edit">
      <Pencil size={16} />
    </button>
  )
}

export function DeleteButton({ onClick, disabled, loading }: { onClick: () => void; disabled?: boolean; loading?: boolean }) {
  return (
    <button
      onClick={onClick}
      disabled={disabled || loading}
      className="p-2 rounded-lg border bg-red-500/20 hover:bg-red-500/30 text-red-400 border-red-500/30 transition-all disabled:opacity-50"
      title="Delete"
    >
      {loading ? (
        <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
      ) : (
        <Trash2 size={16} />
      )}
    </button>
  )
}
