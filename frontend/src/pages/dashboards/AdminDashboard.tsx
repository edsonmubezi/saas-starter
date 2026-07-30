import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import {
  AreaChart,
  Area,
  ResponsiveContainer,
  XAxis,
  YAxis,
  Tooltip,
} from 'recharts'
import {
  Building2,
  Users,
  UserCheck,
  ShieldAlert,
  Lock,
  ChevronRight,
  Activity,
  type LucideIcon,
} from 'lucide-react'
import clsx from 'clsx'

import {
  getAdminDashboard,
  type AdminDashboard,
} from '../../utils/dashboards'

// ─── Color system ───────────────────────────────────────────────────────────
type CardColor = 'blue' | 'green' | 'yellow' | 'red' | 'purple' | 'cyan' | 'indigo' | 'amber'

const COLOR_STYLES: Record<CardColor, { bg: string; border: string; iconBg: string; iconColor: string; valueColor: string }> = {
  blue:   { bg: 'from-blue-500/10 to-blue-600/5',     border: 'border-blue-500/20',   iconBg: 'bg-blue-500/20',   iconColor: 'text-blue-400',   valueColor: 'text-foreground' },
  green:  { bg: 'from-emerald-500/10 to-emerald-600/5', border: 'border-emerald-500/20', iconBg: 'bg-emerald-500/20', iconColor: 'text-emerald-400', valueColor: 'text-emerald-400' },
  yellow: { bg: 'from-yellow-500/10 to-yellow-600/5',   border: 'border-yellow-500/20', iconBg: 'bg-yellow-500/20', iconColor: 'text-yellow-400', valueColor: 'text-yellow-400' },
  red:    { bg: 'from-red-500/10 to-red-600/5',         border: 'border-red-500/20',    iconBg: 'bg-red-500/20',    iconColor: 'text-red-400',    valueColor: 'text-red-400' },
  purple: { bg: 'from-purple-500/10 to-purple-600/5',   border: 'border-purple-500/20', iconBg: 'bg-purple-500/20', iconColor: 'text-purple-400', valueColor: 'text-purple-400' },
  cyan:   { bg: 'from-cyan-500/10 to-cyan-600/5',       border: 'border-cyan-500/20',   iconBg: 'bg-cyan-500/20',   iconColor: 'text-cyan-400',   valueColor: 'text-cyan-400' },
  indigo: { bg: 'from-indigo-500/10 to-indigo-600/5',   border: 'border-indigo-500/20', iconBg: 'bg-indigo-500/20', iconColor: 'text-indigo-400', valueColor: 'text-indigo-400' },
  amber:  { bg: 'from-amber-500/10 to-amber-600/5',     border: 'border-amber-500/20',  iconBg: 'bg-amber-500/20',  iconColor: 'text-amber-400',  valueColor: 'text-amber-400' },
}

// ─── MetricCard ─────────────────────────────────────────────────────────────
function MetricCard({
  label, value, subtext, color = 'blue', icon: Icon, to,
}: {
  label: string; value: string | number; subtext?: string
  color?: CardColor; icon?: LucideIcon; to?: string
}) {
  const s = COLOR_STYLES[color]
  const card = (
    <div className={`bg-gradient-to-br ${s.bg} border ${s.border} rounded-xl p-5 backdrop-blur transition-all hover:scale-[1.02]`}>
      <div className="flex items-center justify-between">
        <div className="min-w-0">
          <p className="text-foreground/60 text-sm font-medium">{label}</p>
          <p className={`text-3xl font-bold ${s.valueColor} mt-1 tabular-nums`}>{value}</p>
          {subtext && <p className="text-foreground/40 text-xs mt-1">{subtext}</p>}
        </div>
        {Icon && (
          <div className={`w-12 h-12 rounded-lg ${s.iconBg} flex items-center justify-center shrink-0`}>
            <Icon className={`w-6 h-6 ${s.iconColor}`} />
          </div>
        )}
      </div>
    </div>
  )
  return to ? <Link to={to} className="block">{card}</Link> : card
}

// ─── ChartTooltip ───────────────────────────────────────────────────────────
const ChartTooltip = ({ active, payload, label, formatter }: any) => {
  if (!active || !payload?.length) return null
  return (
    <div className="bg-surface-primary border border-foreground/10 rounded-lg px-3 py-2 shadow-xl">
      <p className="text-foreground/60 text-xs mb-1">{label}</p>
      <p className="text-foreground font-semibold">{formatter ? formatter(payload[0].value) : payload[0].value}</p>
    </div>
  )
}

// ─── QuickLink ──────────────────────────────────────────────────────────────
function QuickLink({ to, label, icon: Icon, color }: { to: string; label: string; icon: LucideIcon; color: string }) {
  return (
    <Link
      to={to}
      className="flex items-center gap-3 p-4 rounded-xl border border-foreground/5 bg-foreground/[0.02] hover:bg-foreground/[0.06] transition-colors"
    >
      <Icon size={20} className={color} />
      <span className="text-sm font-medium text-foreground/80">{label}</span>
      <ChevronRight size={14} className="ml-auto text-foreground/25" />
    </Link>
  )
}

// ─── SecurityBadge ──────────────────────────────────────────────────────────
function SecurityBadge({ count, label, type = 'warning' }: { count: number; label: string; type?: 'warning' | 'danger' }) {
  const styles = {
    warning: 'bg-amber-500/10 border-amber-500/20 text-amber-400',
    danger: 'bg-red-500/10 border-red-500/20 text-red-400',
  }
  const dotColor = {
    warning: 'bg-amber-400',
    danger: 'bg-red-400',
  }
  return (
    <div className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-full border ${styles[type]} text-sm`}>
      <span className="relative flex h-2 w-2">
        <span className={`animate-ping absolute inline-flex h-full w-full rounded-full ${dotColor[type]} opacity-75`} />
        <span className={`relative inline-flex rounded-full h-2 w-2 ${dotColor[type]}`} />
      </span>
      <span className="font-medium">{count}</span>
      <span className="text-foreground/60">{label}</span>
    </div>
  )
}

// ═════════════════════════════════════════════════════════════════════════════
// MAIN COMPONENT
// ═════════════════════════════════════════════════════════════════════════════
export default function AdminDashboardPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['adminDashboard'],
    queryFn: getAdminDashboard,
    staleTime: 5 * 60 * 1000,
    refetchInterval: 5 * 60 * 1000,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="flex items-center gap-3 text-foreground/60">
          <div className="w-5 h-5 border-2 border-foreground/20 border-t-white/60 rounded-full animate-spin" />
          Loading dashboard...
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="text-red-400 mb-2">Failed to load dashboard</div>
          <button onClick={() => window.location.reload()} className="text-sm text-foreground/60 hover:text-foreground underline">
            Try again
          </button>
        </div>
      </div>
    )
  }

  const d = data as AdminDashboard

  return (
    <div className="space-y-6">

      {/* ─── Header ──────────────────────────────────────────────────────── */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Admin Dashboard</h1>
          <p className="text-foreground/40 text-sm mt-1">System-wide overview across all organizations</p>
        </div>
        <div className="flex flex-wrap gap-2">
          {(d?.failed_login_attempts ?? 0) > 0 && (
            <SecurityBadge count={d.failed_login_attempts} label="failed logins" type="warning" />
          )}
          {(d?.locked_accounts ?? 0) > 0 && (
            <SecurityBadge count={d.locked_accounts} label="locked accounts" type="danger" />
          )}
        </div>
      </div>

      {/* ─── Platform Metrics ─────────────────────────────────────────────── */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <MetricCard
          label="Organizations"
          value={d?.total_organizations ?? 0}
          color="indigo"
          icon={Building2}
          to="/organizations"
        />
        <MetricCard
          label="Total Users"
          value={d?.total_users ?? 0}
          subtext={`${d?.active_users_30_days ?? 0} active in 30 days`}
          color="blue"
          icon={Users}
          to="/users"
        />
        <MetricCard
          label="Active Users"
          value={d?.active_users_30_days ?? 0}
          subtext="last 30 days"
          color="green"
          icon={UserCheck}
        />
      </div>

      {/* ─── Security Stats ─────────────────────────────────────────────── */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div className={clsx(
          'bg-gradient-to-br rounded-xl p-5 border backdrop-blur flex items-center gap-4',
          (d?.failed_login_attempts ?? 0) > 10
            ? 'from-red-500/10 to-red-600/5 border-red-500/20'
            : 'from-foreground/[0.03] to-foreground/[0.01] border-foreground/[0.06]'
        )}>
          <div className={clsx(
            'w-12 h-12 rounded-lg flex items-center justify-center shrink-0',
            (d?.failed_login_attempts ?? 0) > 10 ? 'bg-red-500/20' : 'bg-foreground/[0.06]'
          )}>
            <ShieldAlert className={clsx('w-6 h-6', (d?.failed_login_attempts ?? 0) > 10 ? 'text-red-400' : 'text-foreground/40')} />
          </div>
          <div>
            <p className="text-foreground/60 text-sm font-medium">Failed Login Attempts</p>
            <p className="text-2xl font-bold text-foreground tabular-nums">{d?.failed_login_attempts ?? 0}</p>
          </div>
        </div>

        <div className={clsx(
          'bg-gradient-to-br rounded-xl p-5 border backdrop-blur flex items-center gap-4',
          (d?.locked_accounts ?? 0) > 0
            ? 'from-amber-500/10 to-amber-600/5 border-amber-500/20'
            : 'from-foreground/[0.03] to-foreground/[0.01] border-foreground/[0.06]'
        )}>
          <div className={clsx(
            'w-12 h-12 rounded-lg flex items-center justify-center shrink-0',
            (d?.locked_accounts ?? 0) > 0 ? 'bg-amber-500/20' : 'bg-foreground/[0.06]'
          )}>
            <Lock className={clsx('w-6 h-6', (d?.locked_accounts ?? 0) > 0 ? 'text-amber-400' : 'text-foreground/40')} />
          </div>
          <div>
            <p className="text-foreground/60 text-sm font-medium">Locked Accounts</p>
            <p className="text-2xl font-bold text-foreground tabular-nums">{d?.locked_accounts ?? 0}</p>
          </div>
        </div>
      </div>

      {/* ─── User Growth Chart ──────────────────────────────────────────── */}
      <div className="bg-foreground/[0.02] rounded-xl border border-foreground/5 p-5">
        <div className="flex items-center justify-between mb-6">
          <h2 className="font-semibold text-foreground">User Growth</h2>
          <span className="text-xs text-foreground/40">Last 12 months</span>
        </div>
        <div className="h-56">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={d?.user_growth_trend || []}>
              <defs>
                <linearGradient id="colorUsersGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#6366F1" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#6366F1" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis dataKey="month" axisLine={false} tickLine={false} tick={{ fill: '#6B7280', fontSize: 11 }} dy={10} />
              <YAxis axisLine={false} tickLine={false} tick={{ fill: '#6B7280', fontSize: 11 }} dx={-10} />
              <Tooltip content={<ChartTooltip />} />
              <Area
                type="monotone"
                dataKey="count"
                stroke="#6366F1"
                strokeWidth={2}
                fillOpacity={1}
                fill="url(#colorUsersGrad)"
                name="New Users"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* ─── Quick Access ─────────────────────────────────────────────────── */}
      <div>
        <h3 className="text-xs uppercase tracking-wider text-foreground/40 font-medium mb-3">Quick Access</h3>
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
          <QuickLink to="/organizations" label="Organizations" icon={Building2} color="text-indigo-400" />
          <QuickLink to="/users" label="Users" icon={Users} color="text-blue-400" />
          <QuickLink to="/audit-events" label="Audit Trail" icon={Activity} color="text-cyan-400" />
          <QuickLink to="/security" label="Security" icon={ShieldAlert} color="text-red-400" />
          <QuickLink to="/system-logs" label="System Logs" icon={Activity} color="text-amber-400" />
        </div>
      </div>
    </div>
  )
}
