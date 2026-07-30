import React from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from '../state/AuthContext'

type PermissionRouteProps = {
  children: React.ReactNode
  requiredPermissions?: string[]
  requireAny?: boolean // If true, only one of the permissions is needed; if false, all are required
}

export default function PermissionRoute({
  children,
  requiredPermissions = [],
  requireAny = false
}: PermissionRouteProps) {
  const { user, loading } = useAuth()

  if (loading) {
    return <div className="min-h-screen grid place-items-center text-foreground/80">Loading…</div>
  }

  if (!user) {
    return <Navigate to="/login" replace />
  }

  // If no permissions are required, allow access
  if (requiredPermissions.length === 0) {
    return <>{children}</>
  }

  const userPermissions = new Set<string>(Array.isArray(user?.permissions) ? user.permissions : [])

  let hasAccess = false

  if (requireAny) {
    // User needs at least one of the required permissions
    hasAccess = requiredPermissions.some(permission => userPermissions.has(permission))
  } else {
    // User needs all of the required permissions
    hasAccess = requiredPermissions.every(permission => userPermissions.has(permission))
  }

  if (!hasAccess) {
    return (
      <div className="min-h-screen grid place-items-center bg-surface-primary text-foreground">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-red-400 mb-4">Access Denied</h1>
          <p className="text-foreground/70">You don't have permission to access this page.</p>
          <button
            onClick={() => window.history.back()}
            className="mt-4 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors"
          >
            Go Back
          </button>
        </div>
      </div>
    )
  }

  return <>{children}</>
}