import React, { useEffect, useState } from 'react'
import { getDashboardStats, type Stats } from '../utils/dashboard'
import { AreaChart, Area, ResponsiveContainer, XAxis, YAxis, Tooltip, CartesianGrid } from 'recharts'

export default function DashboardPage() {
  const [stats, setStats] = useState<Stats | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    (async () => {
      try {
        const s = await getDashboardStats()
        setStats(s) // ✅ pass the fetched data
      } catch (e: any) {
        // Fallback demo data if fetching fails
        setStats({
          totals: { users: 1280, orgs: 12, activeUsers: 436 },
          trend: Array.from({ length: 12 }).map((_, i) => ({
            date: `2025-${String(i + 1).padStart(2, '0')}`,
            users: 200 + Math.round(Math.random() * 300),
          })),
        })
        setError(e?.message || 'Failed to load stats')
      }
    })()
  }, [])

  return (
    <div className="grid gap-6">
      {error && <div className="text-sm text-rose-400">Note: showing demo data — {error}</div>}

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="card">
          <div className="card-title">Total Users</div>
          <div className="card-value">{stats?.totals.users ?? 0}</div>
        </div>
        <div className="card">
          <div className="card-title">Organizations</div>
          <div className="card-value">{stats?.totals.orgs ?? 0}</div>
        </div>
        <div className="card">
          <div className="card-title">Active Users (30d)</div>
          <div className="card-value">{stats?.totals.activeUsers ?? 0}</div>
        </div>
      </div>

      <div className="card h-80">
        <div className="flex items-center justify-between mb-4">
          <h2 className="font-semibold">User Growth</h2>
        </div>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={stats?.trend || []}>
              <defs>
                <linearGradient id="colorUsers" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#2563eb" stopOpacity={0.4} />
                  <stop offset="95%" stopColor="#2563eb" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis />
              <Tooltip />
              <Area type="monotone" dataKey="users" stroke="#2563eb" fillOpacity={1} fill="url(#colorUsers)" />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  )
}
