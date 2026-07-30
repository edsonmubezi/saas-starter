import React from 'react'
import { Link } from 'react-router-dom'
import { Plus, ArrowLeft } from 'lucide-react'

interface PageHeaderProps {
  title: string
  subtitle?: string
  /** Back button - provide path to show back arrow */
  backTo?: string
  /** Primary action button */
  action?: {
    label: string
    to?: string
    onClick?: () => void
    icon?: React.ReactNode
  }
  /** Secondary actions */
  secondaryActions?: React.ReactNode
  /** Additional content below title (e.g., badges, status) */
  meta?: React.ReactNode
}

export default function PageHeader({
  title,
  subtitle,
  backTo,
  action,
  secondaryActions,
  meta,
}: PageHeaderProps) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
      <div className="flex items-start gap-3">
        {backTo && (
          <Link
            to={backTo}
            className="mt-1 p-2 rounded-lg bg-foreground/5 hover:bg-foreground/10 border border-foreground/10 text-foreground/60 hover:text-foreground transition-all"
          >
            <ArrowLeft size={18} />
          </Link>
        )}
        <div>
          <h1 className="text-xl md:text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            {title}
          </h1>
          {subtitle && (
            <p className="text-foreground/60 text-sm mt-1">{subtitle}</p>
          )}
          {meta && <div className="mt-2">{meta}</div>}
        </div>
      </div>

      <div className="flex items-center gap-3">
        {secondaryActions}
        {action && (
          action.to ? (
            <Link
              to={action.to}
              className="px-4 py-2.5 rounded-xl bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-500 hover:to-blue-600 text-white font-medium shadow-lg shadow-blue-500/20 transition-all flex items-center gap-2"
            >
              {action.icon || <Plus size={18} />}
              {action.label}
            </Link>
          ) : (
            <button
              onClick={action.onClick}
              className="px-4 py-2.5 rounded-xl bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-500 hover:to-blue-600 text-white font-medium shadow-lg shadow-blue-500/20 transition-all flex items-center gap-2"
            >
              {action.icon || <Plus size={18} />}
              {action.label}
            </button>
          )
        )}
      </div>
    </div>
  )
}
