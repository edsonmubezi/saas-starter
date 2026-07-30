import { useEffect, useState } from 'react'
import { useAuth } from '../state/AuthContext'
import AdminDashboard from './dashboards/AdminDashboard'
import TenantDashboard from './dashboards/TenantDashboard'
import PageLoader from '../ui/PageLoader'

export default function DashboardRouter() {
  const { user } = useAuth()
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    // Small delay to ensure permissions are loaded
    const timer = setTimeout(() => setIsLoading(false), 100)
    return () => clearTimeout(timer)
  }, [])

  if (isLoading) {
    return <PageLoader />
  }

  // Check user permissions to determine which dashboard to show
  const permissions = user?.permissions || []

  // SuperAdmin gets admin dashboard
  if (permissions.some(p => p.startsWith('admin.dashboard.') || p === 'admin.*')) {
    return <AdminDashboard />
  }

  // OrgAdmin/Tenant gets tenant dashboard
  return <TenantDashboard />
}
