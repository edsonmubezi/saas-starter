import { request } from './common'
import { API_PREFIX } from './apiPrefix'

// ============================================================================
// Admin Dashboard Types
// ============================================================================

export interface MonthlyCount {
  month: string
  count: number
}

export interface AdminDashboard {
  // Organization Overview
  total_organizations: number
  total_users: number
  active_users_30_days: number

  // Security & Compliance
  failed_login_attempts: number
  locked_accounts: number

  // Trends
  organization_growth_trend: MonthlyCount[]
  user_growth_trend: MonthlyCount[]
}

// ============================================================================
// API Functions
// ============================================================================

/**
 * Fetch Admin Dashboard data (SuperAdmin/HQ level)
 */
export async function getAdminDashboard(): Promise<AdminDashboard> {
  const res = await request<{ data: AdminDashboard }>(`${API_PREFIX.ADMIN}/dashboard`)
  return res.data
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Format currency for display
 */
export function formatCurrency(amount: number, currency: string = 'TZS'): string {
  return new Intl.NumberFormat('en-TZ', {
    style: 'currency',
    currency: currency,
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
}

/**
 * Format percentage for display
 */
export function formatPercentage(value: number, decimals: number = 1): string {
  return `${value.toFixed(decimals)}%`
}
