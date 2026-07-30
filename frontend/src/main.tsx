import React, { Suspense, lazy } from 'react'
import ReactDOM from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import './index.css'
import { AuthProvider } from './state/AuthContext'
import ErrorBoundary from './ui/ErrorBoundary'
import AppLayout from './ui/AppLayout'
import ProtectedRoute from './ui/ProtectedRoute'
import PermissionRoute from './ui/PermissionRoute'
import PageLoader from './ui/PageLoader'
import { Toaster } from 'react-hot-toast'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfirmProvider } from "./ui/ConfirmProvider"
import { ThemeProvider } from './state/ThemeContext'
import { WebSocketProvider } from './state/WebSocketContext'

// ============================================================================
// Global handler for chunk loading failures (stale deployments)
// ============================================================================
window.addEventListener('error', (event) => {
  const message = event.message || ''
  if (
    message.includes('Failed to fetch dynamically imported module') ||
    message.includes('Loading chunk') ||
    message.includes('Loading CSS chunk')
  ) {
    const lastReload = sessionStorage.getItem('chunk-error-reload')
    const now = Date.now()
    if (!lastReload || now - parseInt(lastReload) > 10000) {
      sessionStorage.setItem('chunk-error-reload', now.toString())
      window.location.reload()
    }
  }
})

window.addEventListener('unhandledrejection', (event) => {
  const message = event.reason?.message || String(event.reason) || ''
  if (
    message.includes('Failed to fetch dynamically imported module') ||
    message.includes('Loading chunk') ||
    message.includes('Loading CSS chunk')
  ) {
    const lastReload = sessionStorage.getItem('chunk-error-reload')
    const now = Date.now()
    if (!lastReload || now - parseInt(lastReload) > 10000) {
      sessionStorage.setItem('chunk-error-reload', now.toString())
      window.location.reload()
    }
  }
})

// ============================================================================
// Lazy-loaded page components (code-splitting for better initial load)
// ============================================================================

// Auth
const LoginPage = lazy(() => import('./pages/Login'))
const TwoFactorVerifyPage = lazy(() => import('./pages/TwoFactorVerify'))
const ForgotPasswordPage = lazy(() => import('./pages/auth/ForgotPasswordPage'))
const ResetPasswordPage = lazy(() => import('./pages/auth/ResetPasswordPage'))
const DashboardRouter = lazy(() => import('./pages/DashboardRouter'))

// Users
const UsersPage = lazy(() => import('./pages/users/Users'))
const SuperAdminUsersPage = lazy(() => import('./pages/users/SuperAdminUsers'))
const OrgAdminUsersPage = lazy(() => import('./pages/users/OrgAdminUsers'))
const OrgUserDetailPage = lazy(() => import('./pages/users/UserDetailPage'))
const HQAssignUserRoles = lazy(() => import('./pages/users/HQAssignUserRoles'))
const OrgAssignRoles = lazy(() => import('./pages/users/OrgAssignRoles'))
const OrgCreateRole = lazy(() => import('./pages/users/OrgCreateRole'))

// Organizations
const OrgsPage = lazy(() => import('./pages/organisation/Orgs'))
const OrgSettingsPage = lazy(() => import('./pages/organisation/OrgSettings'))
const DocumentBrandingPage = lazy(() => import('./pages/organisation/DocumentBrandingPage'))
const EmailBrandingPage = lazy(() => import('./pages/organisation/EmailBrandingPage'))

// Authorization
const RolesPage = lazy(() => import('./pages/Authorize/RolesPage'))
const PermissionsPage = lazy(() => import('./pages/Authorize/PermissionsPage'))
const AssignRolePermissionsPageMock = lazy(() => import('./pages/Authorize/AssignRolePermissionsPage'))
const NonHqAssignPermission = lazy(() => import('./pages/Authorize/NonHqAssignPermission'))

// Settings
const SecuritySettingsPage = lazy(() => import('./pages/settings/SecuritySettings'))

// Alerting
const AlertConfigsPage = lazy(() => import('./pages/Alerting/AlertConfigs'))
const AlertHistoryPage = lazy(() => import('./pages/Alerting/AlertHistory'))

// Audit
const AuditEventsPage = lazy(() => import('./pages/Audit/AuditEvents'))
// Security
const SecurityDashboardPage = lazy(() => import('./pages/Security/SecurityDashboard'))
const SecurityEventsPage = lazy(() => import('./pages/Security/SecurityEvents'))

// Logs
const ApplicationLogsPage = lazy(() => import('./pages/Logs/ApplicationLogs'))
const AccessLogsPage = lazy(() => import('./pages/Logs/AccessLogs'))

// Dashboards
const AdminDashboard = lazy(() => import('./pages/dashboards/AdminDashboard'))
const TenantDashboard = lazy(() => import('./pages/dashboards/TenantDashboard'))

// Knowledge Base
const KnowledgeBasePage = lazy(() => import('./pages/Chat/KnowledgeBasePage'))

// ============================================================================
// Suspense wrapper for lazy components with error handling
// ============================================================================
const LazyPage = ({ children }: { children: React.ReactNode }) => (
  <ErrorBoundary>
    <Suspense fallback={<PageLoader />}>{children}</Suspense>
  </ErrorBoundary>
)

// ============================================================================
// Router configuration
// ============================================================================
const router = createBrowserRouter([
  { path: '/login', element: <LazyPage><LoginPage /></LazyPage> },
  { path: '/2fa-verify', element: <LazyPage><TwoFactorVerifyPage /></LazyPage> },
  { path: '/forgot-password', element: <LazyPage><ForgotPasswordPage /></LazyPage> },
  { path: '/reset-password', element: <LazyPage><ResetPasswordPage /></LazyPage> },
  {
    path: '/',
    element: (
      <ProtectedRoute>
        <AppLayout />
      </ProtectedRoute>
    ),
    children: [
      { index: true, element: <LazyPage><DashboardRouter /></LazyPage> },

      // Dashboards
      { path: 'dashboard/admin', element: <PermissionRoute requiredPermissions={['admin.dashboard.view']}><LazyPage><AdminDashboard /></LazyPage></PermissionRoute> },
      { path: 'dashboard/tenant', element: <PermissionRoute requiredPermissions={['tenant.dashboard.view']}><LazyPage><TenantDashboard /></LazyPage></PermissionRoute> },

      // Users
      { path: 'users', element: <PermissionRoute requiredPermissions={['admin.user.view']}><LazyPage><UsersPage /></LazyPage></PermissionRoute> },
      { path: 'superadmin-users', element: <PermissionRoute requiredPermissions={['admin.user.view']}><LazyPage><SuperAdminUsersPage /></LazyPage></PermissionRoute> },
      { path: 'org-users', element: <PermissionRoute requiredPermissions={['tenant.user.view']}><LazyPage><OrgAdminUsersPage /></LazyPage></PermissionRoute> },
      { path: 'org-users/:id', element: <PermissionRoute requiredPermissions={['tenant.user.view']}><LazyPage><OrgUserDetailPage /></LazyPage></PermissionRoute> },
      { path: 'assign-user-roles', element: <PermissionRoute requiredPermissions={['admin.role.assign']}><LazyPage><HQAssignUserRoles /></LazyPage></PermissionRoute> },
      { path: 'org-assign-roles', element: <PermissionRoute requiredPermissions={['tenant.role.assign']}><LazyPage><OrgAssignRoles /></LazyPage></PermissionRoute> },
      { path: 'org-create-role', element: <PermissionRoute requiredPermissions={['tenant.role.create']}><LazyPage><OrgCreateRole /></LazyPage></PermissionRoute> },

      // Organizations
      { path: 'view-organisations', element: <PermissionRoute requiredPermissions={['admin.organization.view']}><LazyPage><OrgsPage /></LazyPage></PermissionRoute> },
      { path: 'organisations/:id/settings', element: <PermissionRoute requiredPermissions={['admin.org_settings.view']}><LazyPage><OrgSettingsPage /></LazyPage></PermissionRoute> },

      // Authorization
      { path: 'manage-roles', element: <PermissionRoute requiredPermissions={['admin.role.view']}><LazyPage><RolesPage /></LazyPage></PermissionRoute> },
      { path: 'view-permission', element: <PermissionRoute requiredPermissions={['admin.permission.view']}><LazyPage><PermissionsPage /></LazyPage></PermissionRoute> },
      { path: 'assign-role-permits', element: <PermissionRoute requiredPermissions={['admin.permission.edit']}><LazyPage><AssignRolePermissionsPageMock /></LazyPage></PermissionRoute> },
      { path: 'assign-permission', element: <PermissionRoute requiredPermissions={['admin.permission.viewltd']}><LazyPage><NonHqAssignPermission /></LazyPage></PermissionRoute> },

      // Document Branding
      { path: 'document-branding', element: <PermissionRoute requiredPermissions={['tenant.document_branding.manage']}><LazyPage><DocumentBrandingPage /></LazyPage></PermissionRoute> },
      { path: 'email-branding', element: <PermissionRoute requiredPermissions={['tenant.email_branding.manage']}><LazyPage><EmailBrandingPage /></LazyPage></PermissionRoute> },

      // Settings (accessible to all authenticated users)
      { path: 'settings/security', element: <LazyPage><SecuritySettingsPage /></LazyPage> },

      // Alerting (accessible to admins)
      { path: 'alerting/configs', element: <LazyPage><AlertConfigsPage /></LazyPage> },
      { path: 'alerting/history', element: <LazyPage><AlertHistoryPage /></LazyPage> },

      // Audit (accessible to admins)
      { path: 'audit/events', element: <LazyPage><AuditEventsPage /></LazyPage> },
      // Security (accessible to admins)
      { path: 'security/dashboard', element: <LazyPage><SecurityDashboardPage /></LazyPage> },
      { path: 'security/events', element: <LazyPage><SecurityEventsPage /></LazyPage> },

      // Application Logs (accessible to admins)
      { path: 'logs/application', element: <LazyPage><ApplicationLogsPage /></LazyPage> },
      { path: 'logs/access', element: <LazyPage><AccessLogsPage /></LazyPage> },

      // Knowledge Base
      { path: 'knowledge-base', element: <PermissionRoute requiredPermissions={['tenant.knowledge_base.view']}><LazyPage><KnowledgeBasePage /></LazyPage></PermissionRoute> },
    ]
  },
])

// ============================================================================
// React Query configuration with optimized defaults
// ============================================================================
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,      // 5 minutes - data stays fresh
      gcTime: 1000 * 60 * 10,         // 10 minutes - cache garbage collection (formerly cacheTime)
      retry: (failureCount, error: any) => {
        // Never retry auth errors — bail out immediately
        if (error?.status === 401 || error?.isAuthExpired) return false
        return failureCount < 1 // retry once for other errors
      },
      refetchOnWindowFocus: false,    // Don't refetch when window regains focus
      refetchOnReconnect: true,       // Refetch when connection is restored
    },
    mutations: {
      retry: 0,                        // Don't retry mutations
    },
  },
})

// Cancel all queries when auth expires — prevents stale requests from hanging the UI
window.addEventListener('auth:expired', () => {
  queryClient.cancelQueries()
  queryClient.clear()
})

// ============================================================================
// App render
// ============================================================================
ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ErrorBoundary>
      <ThemeProvider>
        <AuthProvider>
          <QueryClientProvider client={queryClient}>
            <WebSocketProvider>
              <ConfirmProvider>
                <RouterProvider router={router} />
                <Toaster position="top-right" />
              </ConfirmProvider>
            </WebSocketProvider>
          </QueryClientProvider>
        </AuthProvider>
      </ThemeProvider>
    </ErrorBoundary>
  </React.StrictMode>
)
