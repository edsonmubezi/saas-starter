import React from 'react'
import { useAuth } from '../../state/AuthContext'
import { LayoutDashboard } from 'lucide-react'

// ═══════════════════════════════════════════════════════════════════════════════
// TENANT DASHBOARD - Organization-level overview
// ═══════════════════════════════════════════════════════════════════════════════
export default function TenantDashboardPage() {
  const { user } = useAuth()

  return (
    <div className="space-y-6">
      {/* ─── Header ──────────────────────────────────────────────────────── */}
      <div>
        <h1 className="text-2xl font-bold text-foreground">Dashboard</h1>
        <p className="text-foreground/40 text-sm mt-1">
          Welcome{user?.organization?.name ? ` to ${user.organization.name}` : ''}
        </p>
      </div>

      {/* ─── Welcome Card ────────────────────────────────────────────────── */}
      <div className="bg-gradient-to-br from-blue-500/10 to-indigo-600/5 border border-blue-500/20 rounded-xl p-8 backdrop-blur">
        <div className="flex items-center gap-4">
          <div className="w-14 h-14 rounded-lg bg-blue-500/20 flex items-center justify-center shrink-0">
            <LayoutDashboard className="w-7 h-7 text-blue-400" />
          </div>
          <div>
            <h2 className="text-xl font-semibold text-foreground">
              {user?.organization?.name || 'Organization'} Dashboard
            </h2>
            <p className="text-foreground/50 text-sm mt-1">
              Your organization dashboard is ready. Features and modules will appear here as they are configured.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
