import React, { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Shield, AlertTriangle, Lock, Activity, Globe, TrendingUp } from 'lucide-react'
import { useAuth } from '../../state/AuthContext'

import {
  getSecurityDashboard,
  type SecurityDashboard,
  SECURITY_SEVERITIES,
  SECURITY_CATEGORIES,
} from '../../utils/security'

export default function SecurityDashboardPage() {
  const { user } = useAuth()
  const perms: string[] = Array.isArray((user as any)?.permissions) ? (user as any).permissions : []
  const canView = perms.some(p => p.includes('admin') || p.includes('security'))

  const [hours, setHours] = useState(24)

  const { data: dashboard, isLoading } = useQuery<SecurityDashboard>({
    queryKey: ['securityDashboard', hours],
    queryFn: () => getSecurityDashboard(hours),
    enabled: canView,
    refetchInterval: 60000, // Refresh every minute
  })

  if (!canView) {
    return (
      <div className="card p-6 bg-foreground/5 border-foreground/10">
        <div className="text-sm text-foreground/80 font-medium">
          You do not have permission to view the security dashboard.
        </div>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="grid gap-4">
        <div className="h-8 bg-foreground/5 rounded animate-pulse w-48" />
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map(i => (
            <div key={i} className="card bg-foreground/5 border-foreground/10 h-32 animate-pulse" />
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="grid gap-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold flex items-center gap-2">
            <Shield className="text-brand-400" />
            Security Dashboard
          </h1>
          <p className="text-sm text-foreground/60 mt-1">
            Monitor security events and potential threats
          </p>
        </div>
        <div className="flex items-center gap-2">
          <label className="text-sm text-foreground/60">Time Range:</label>
          <select
            value={hours}
            onChange={(e) => setHours(Number(e.target.value))}
            className="input w-auto"
          >
            <option value={1}>Last 1 hour</option>
            <option value={6}>Last 6 hours</option>
            <option value={24}>Last 24 hours</option>
            <option value={72}>Last 3 days</option>
            <option value={168}>Last 7 days</option>
            <option value={720}>Last 30 days</option>
          </select>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="card bg-foreground/5 border-foreground/10">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm text-foreground/60">Total Events</p>
              <p className="text-3xl font-bold mt-1">{dashboard?.total_events || 0}</p>
              <p className="text-xs text-foreground/40 mt-1">Last {hours}h</p>
            </div>
            <div className="p-2 bg-blue-500/10 rounded-lg">
              <Activity size={24} className="text-blue-400" />
            </div>
          </div>
        </div>

        <div className="card bg-foreground/5 border-foreground/10">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm text-foreground/60">Failed Logins</p>
              <p className="text-3xl font-bold mt-1 text-amber-400">{dashboard?.failed_logins || 0}</p>
              <p className="text-xs text-foreground/40 mt-1">Authentication failures</p>
            </div>
            <div className="p-2 bg-amber-500/10 rounded-lg">
              <Lock size={24} className="text-amber-400" />
            </div>
          </div>
        </div>

        <div className="card bg-foreground/5 border-foreground/10">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm text-foreground/60">Suspicious Activities</p>
              <p className="text-3xl font-bold mt-1 text-rose-400">{dashboard?.suspicious_activities || 0}</p>
              <p className="text-xs text-foreground/40 mt-1">Potential threats</p>
            </div>
            <div className="p-2 bg-rose-500/10 rounded-lg">
              <AlertTriangle size={24} className="text-rose-400" />
            </div>
          </div>
        </div>

        <div className="card bg-foreground/5 border-foreground/10">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm text-foreground/60">Unique IPs</p>
              <p className="text-3xl font-bold mt-1">{dashboard?.top_ip_addresses?.length || 0}</p>
              <p className="text-xs text-foreground/40 mt-1">Active sources</p>
            </div>
            <div className="p-2 bg-emerald-500/10 rounded-lg">
              <Globe size={24} className="text-emerald-400" />
            </div>
          </div>
        </div>
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Events by Severity */}
        <div className="card bg-foreground/5 border-foreground/10">
          <h3 className="text-sm font-medium mb-4 flex items-center gap-2">
            <TrendingUp size={16} className="text-foreground/40" />
            Events by Severity
          </h3>
          <div className="space-y-3">
            {SECURITY_SEVERITIES.map(sev => {
              const count = dashboard?.events_by_severity?.[sev.value] || 0
              const total = dashboard?.total_events || 1
              const percentage = Math.round((count / total) * 100) || 0
              return (
                <div key={sev.value}>
                  <div className="flex justify-between text-sm mb-1">
                    <span className={sev.color}>{sev.label}</span>
                    <span className="text-foreground/60">{count}</span>
                  </div>
                  <div className="h-2 bg-foreground/5 rounded-full overflow-hidden">
                    <div
                      className={`h-full ${sev.bgColor} transition-all`}
                      style={{ width: `${percentage}%` }}
                    />
                  </div>
                </div>
              )
            })}
          </div>
        </div>

        {/* Events by Category */}
        <div className="card bg-foreground/5 border-foreground/10">
          <h3 className="text-sm font-medium mb-4 flex items-center gap-2">
            <Shield size={16} className="text-foreground/40" />
            Events by Category
          </h3>
          <div className="space-y-3">
            {SECURITY_CATEGORIES.map(cat => {
              const count = dashboard?.events_by_category?.[cat.value] || 0
              const total = dashboard?.total_events || 1
              const percentage = Math.round((count / total) * 100) || 0
              return (
                <div key={cat.value}>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="flex items-center gap-2">
                      <span>{cat.icon}</span>
                      <span>{cat.label}</span>
                    </span>
                    <span className="text-foreground/60">{count}</span>
                  </div>
                  <div className="h-2 bg-foreground/5 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-brand-500/50 transition-all"
                      style={{ width: `${percentage}%` }}
                    />
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      </div>

      {/* Bottom Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Recent Threats */}
        <div className="card bg-foreground/5 border-foreground/10">
          <h3 className="text-sm font-medium mb-4 flex items-center gap-2">
            <AlertTriangle size={16} className="text-rose-400" />
            Recent Threats
          </h3>
          {dashboard?.recent_threats?.length ? (
            <div className="space-y-2 max-h-[300px] overflow-y-auto">
              {dashboard.recent_threats.slice(0, 10).map((threat, i) => {
                const sevInfo = SECURITY_SEVERITIES.find(s => s.value === threat.severity)
                return (
                  <div key={i} className="p-3 bg-foreground/5 rounded-lg">
                    <div className="flex items-start justify-between">
                      <div>
                        <div className="font-medium text-sm">{threat.event_type}</div>
                        <div className="text-xs text-foreground/50 mt-0.5">{threat.description}</div>
                      </div>
                      <span className={`px-2 py-0.5 rounded text-xs ${sevInfo?.color} ${sevInfo?.bgColor}`}>
                        {sevInfo?.label}
                      </span>
                    </div>
                    <div className="flex items-center gap-4 mt-2 text-xs text-foreground/40">
                      <span>{threat.ip_address || '—'}</span>
                      <span>{threat.created_at ? new Date(threat.created_at).toLocaleString() : '—'}</span>
                    </div>
                  </div>
                )
              })}
            </div>
          ) : (
            <div className="text-center py-8 text-foreground/40">
              No threats detected in this period
            </div>
          )}
        </div>

        {/* Top IP Addresses */}
        <div className="card bg-foreground/5 border-foreground/10">
          <h3 className="text-sm font-medium mb-4 flex items-center gap-2">
            <Globe size={16} className="text-foreground/40" />
            Top IP Addresses
          </h3>
          {dashboard?.top_ip_addresses?.length ? (
            <div className="space-y-2 max-h-[300px] overflow-y-auto">
              {dashboard.top_ip_addresses.slice(0, 10).map((item, i) => (
                <div key={i} className="flex items-center justify-between p-3 bg-foreground/5 rounded-lg">
                  <span className="font-mono text-sm">{item.ip}</span>
                  <span className="text-foreground/60 text-sm">{item.count} events</span>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-foreground/40">
              No IP data available
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
