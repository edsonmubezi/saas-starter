import React, { useState, useRef, useEffect, lazy, Suspense } from 'react'
import { NavLink, Outlet, useNavigate, useLocation, Link } from 'react-router-dom'
import { logout } from '../utils/api'
import { useAuth } from '../state/AuthContext'
import { clearTokens } from '../state/tokenStore'
import { IdleTimeoutProvider } from '../state/IdleTimeoutProvider'
import { useTheme } from '../state/ThemeContext'
import {
  ChevronDown, LayoutDashboard, LogOut, HelpCircle, Menu, X,
  Settings, Building2, Shield, Clock, ExternalLink,
  Sparkles, MessageSquare, Sun, Moon
} from 'lucide-react'
import clsx from 'clsx'
import toast from 'react-hot-toast'
import HQAdminMenu from './menus/HQAdminMenu'
import OrgAdminMenu from './menus/OrgAdminMenu'
import { useMediaQuery } from '../hooks/useMediaQuery'
import NotificationBell from './NotificationBell'

const ChatWidget = lazy(() => import('./ChatBot/ChatWidget'))

// Get time-based greeting
function getGreeting() {
  const hour = new Date().getHours()
  if (hour < 12) return 'Good morning'
  if (hour < 17) return 'Good afternoon'
  return 'Good evening'
}

function NavItem({ to, children }: { to: string; children: React.ReactNode }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        clsx(
          'block rounded-lg px-3 py-2.5 text-sm text-foreground/80 hover:bg-foreground/5 hover:text-foreground transition-all',
          isActive && 'bg-gradient-to-r from-emerald-500/10 to-teal-500/10 text-foreground font-medium border-l-2 border-emerald-500'
        )
      }
      end
    >
      {children}
    </NavLink>
  )
}

function usePerms(user: any) {
  const perms = new Set<string>(Array.isArray(user?.permissions) ? user.permissions : [])

  const allow = (requiredAny?: string[], requiredAll?: string[]) => {
    if (!perms.size) return false
    if (requiredAll?.length && !requiredAll.every(p => perms.has(p))) return false
    if (requiredAny?.length && !requiredAny.some(p => perms.has(p))) return false
    return true
  }

  const ShowIf = ({
    any, all, children
  }: { any?: string[]; all?: string[]; children: React.ReactNode }) =>
    allow(any, all) ? <>{children}</> : null

  return { allow, ShowIf }
}

// Maps route path prefixes to sidebar section IDs
const ROUTE_TO_SECTION: [string, string][] = [
  // Org Admin sections
  ['/org-users', 'user-management'],
  ['/org-assign-roles', 'user-management'],
  ['/org-create-role', 'user-management'],
  ['/email-config', 'communication'],
  ['/document-branding', 'communication'],
  ['/email-branding', 'communication'],
  ['/knowledge-base', 'communication'],
  // HQ Admin sections
  ['/admin-users', 'hq-users'],
  ['/organizations', 'hq-orgs'],
  ['/admin-security', 'hq-security'],
  ['/admin-audit', 'hq-audit'],
  ['/admin-logs', 'hq-logs'],
]

function getSectionForPath(pathname: string): string | null {
  for (const [prefix, section] of ROUTE_TO_SECTION) {
    if (pathname === prefix || pathname.startsWith(prefix + '/')) return section
  }
  return null
}

export default function AppLayout() {
  const nav = useNavigate()
  const location = useLocation()
  const { user, refresh } = useAuth()
  const { theme, toggleTheme } = useTheme()
  const [menuOpen, setMenuOpen] = useState(false)
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [openSection, setOpenSection] = useState<string | null>(null)
  const menuRef = useRef<HTMLDivElement>(null)
  const { allow, ShowIf } = usePerms(user)
  const isMobile = useMediaQuery('(max-width: 1023px)')
  const toggleSection = (sectionId: string) => {
    setOpenSection(current => current === sectionId ? null : sectionId)
  }

  // Auto-expand the sidebar section matching the current route
  useEffect(() => {
    const section = getSectionForPath(location.pathname)
    if (section) setOpenSection(section)
  }, [location.pathname])

  useEffect(() => {
    const onDocClick = (e: MouseEvent) => {
      if (!menuRef.current) return
      if (!menuRef.current.contains(e.target as Node)) setMenuOpen(false)
    }
    document.addEventListener('click', onDocClick)
    return () => document.removeEventListener('click', onDocClick)
  }, [])

  // Close sidebar on mobile when navigating
  useEffect(() => {
    if (isMobile) setSidebarOpen(false)
  }, [location.pathname])

  async function onLogout() {
    try { await logout() } catch { /* ignore - token may already be invalid */ }
    clearTokens()
    toast.success('Logged out successfully')
    await refresh()
    nav('/login', { replace: true })
  }

  const isLoadingUser = !user || !Array.isArray(user.permissions)

  if (isLoadingUser) {
    return (
      <div className="min-h-screen grid place-items-center bg-surface-primary text-foreground">
        <div className="flex flex-col items-center gap-4">
          <div className="w-10 h-10 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
          <span className="text-foreground/60 text-sm">Loading your workspace...</span>
        </div>
      </div>
    )
  }

  const isHQAdmin = allow(['admin.user.view', 'admin.role.assign'])
  const isOrgAdmin = !isHQAdmin && allow(['tenant.user.view', 'tenant.role.assign'])

  return (
    <IdleTimeoutProvider enabled={true}>
      <div className="min-h-screen bg-surface-primary text-foreground lg:grid lg:grid-cols-[280px_1fr]">

        {/* Mobile Header Bar */}
        <div className="lg:hidden fixed top-0 left-0 right-0 z-50 h-14 bg-surface-secondary/95 backdrop-blur-lg border-b border-foreground/5 flex items-center justify-between px-4">
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-2 rounded-lg hover:bg-foreground/5 transition-all"
            aria-label="Toggle menu"
          >
            {sidebarOpen ? <X size={22} /> : <Menu size={22} />}
          </button>
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center">
              <Sparkles size={16} className="text-white" />
            </div>
            <span className="font-semibold text-sm">SaaS Starter</span>
          </div>
          <div className="w-10" />
        </div>

        {/* Mobile Backdrop */}
        {isMobile && sidebarOpen && (
          <div
            className="fixed inset-0 bg-black/60 backdrop-blur-sm z-30 lg:hidden"
            onClick={() => setSidebarOpen(false)}
          />
        )}

        {/* ========== SIDEBAR ========== */}
        <aside className={clsx(
          "flex flex-col h-screen bg-surface-secondary",
          "fixed inset-y-0 left-0 z-40 w-[280px] transform transition-transform duration-300 ease-out",
          "lg:sticky lg:top-0 lg:translate-x-0 border-r border-foreground/5",
          sidebarOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0"
        )}>

          {/* Logo Section */}
          <div className="px-5 py-5 border-b border-foreground/5">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center shadow-lg shadow-emerald-500/20">
                <Sparkles size={20} className="text-white" />
              </div>
              <div>
                <h1 className="font-bold text-lg tracking-tight">SaaS Starter</h1>
                <p className="text-[10px] text-foreground/40 uppercase tracking-wider">Admin Platform</p>
              </div>
            </div>
          </div>

          {/* Navigation */}
          <div className="flex-1 overflow-y-auto custom-scrollbar px-3 py-4">

            <nav className="space-y-1">
              <NavItem to="/">
                <span className="inline-flex items-center gap-3">
                  <LayoutDashboard size={18} className="text-foreground/60" />
                  <span>Dashboard</span>
                </span>
              </NavItem>

              {isHQAdmin && (
                <HQAdminMenu
                  allow={allow}
                  ShowIf={ShowIf}
                  openSection={openSection}
                  toggleSection={toggleSection}
                />
              )}

              {isOrgAdmin && (
                <OrgAdminMenu
                  allow={allow}
                  ShowIf={ShowIf}
                  openSection={openSection}
                  toggleSection={toggleSection}
                />
              )}
            </nav>

            {/* Quick Links */}
            <div className="mt-8 pt-6 border-t border-foreground/5">
              <p className="px-3 mb-3 text-[10px] uppercase tracking-wider text-foreground/30 font-medium">Support</p>
              <nav className="space-y-1">
                <a
                  href="https://support.ecclesiahr.com"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm text-foreground/60 hover:bg-foreground/5 hover:text-foreground transition-all"
                >
                  <HelpCircle size={18} />
                  <span>Help Center</span>
                  <ExternalLink size={12} className="ml-auto opacity-50" />
                </a>
                <a
                  href="mailto:support@ecclesiahr.com"
                  className="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm text-foreground/60 hover:bg-foreground/5 hover:text-foreground transition-all"
                >
                  <MessageSquare size={18} />
                  <span>Contact Support</span>
                </a>
              </nav>
            </div>
          </div>

          {/* Sidebar Footer - User Profile */}
          <div className="p-4 border-t border-foreground/5 bg-foreground/[0.02]">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500/20 to-teal-500/20 border border-foreground/10 flex items-center justify-center">
                <span className="text-sm font-semibold text-foreground">
                  {(user?.fullname || user?.email || '?').slice(0, 1).toUpperCase()}
                </span>
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-foreground truncate">{user?.fullname || '—'}</p>
                <p className="text-xs text-foreground/40 truncate">{user?.role?.name ?? 'User'}</p>
              </div>
              <button
                onClick={onLogout}
                className="p-2 rounded-lg text-foreground/40 hover:text-red-400 hover:bg-red-500/10 transition-all"
                title="Sign out"
              >
                <LogOut size={18} />
              </button>
            </div>
          </div>
        </aside>

        {/* ========== MAIN CONTENT ========== */}
        <div className="flex flex-col min-h-screen pt-[56px] lg:pt-0">
          {/* ========== HEADER ========== */}
          <header className="mx-4 md:mx-6 mt-4 mb-6 rounded-2xl bg-gradient-to-r from-surface-secondary via-surface-elevated to-surface-secondary border border-foreground/5 p-4 md:p-5">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">

                {/* Left: User Welcome */}
                <div className="flex items-center gap-4">
                  <div className="hidden sm:flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-600 shadow-lg shadow-emerald-500/20">
                    <span className="text-2xl font-bold text-white">
                      {(user?.fullname || user?.email || '?').slice(0, 1).toUpperCase()}
                    </span>
                  </div>
                  <div>
                    <div className="flex items-center gap-2 text-foreground/50 text-xs mb-1">
                      <Clock size={12} />
                      <span>{getGreeting()}</span>
                    </div>
                    <h2 className="text-xl md:text-2xl font-bold text-foreground">
                      {user?.fullname || user?.email}
                    </h2>
                    <div className="flex flex-wrap items-center gap-2 mt-2">
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-medium">
                        <Shield size={12} />
                        {user?.role?.name ?? 'User'}
                      </span>
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-teal-500/10 border border-teal-500/20 text-teal-400 text-xs font-medium">
                        <Building2 size={12} />
                        {user?.organization?.name ?? 'Organization'}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Right: Quick Actions */}
                <div className="flex items-center gap-2 sm:gap-3">
                  {/* Notification Bell */}
                  <NotificationBell />

                  {/* Theme Toggle */}
                  <button
                    onClick={toggleTheme}
                    className="p-3 rounded-xl bg-foreground/5 hover:bg-foreground/10 border border-foreground/10 hover:border-foreground/20 transition-all group"
                    title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
                  >
                    {theme === 'dark' ? (
                      <Sun size={20} className="text-foreground/60 group-hover:text-yellow-400 transition-colors" />
                    ) : (
                      <Moon size={20} className="text-foreground/60 group-hover:text-emerald-500 transition-colors" />
                    )}
                  </button>

                  {/* User Dropdown */}
                  <div className="relative" ref={menuRef}>
                    <button
                      className="flex items-center gap-2 rounded-xl px-3 py-2 bg-foreground/5 hover:bg-foreground/10 border border-foreground/10 hover:border-foreground/20 transition-all"
                      onClick={() => setMenuOpen(o => !o)}
                      aria-haspopup="menu"
                      aria-expanded={menuOpen}
                    >
                      <div className="h-9 w-9 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center">
                        <span className="text-sm font-bold text-white">
                          {(user?.fullname || user?.email || '?').slice(0, 1).toUpperCase()}
                        </span>
                      </div>
                      <ChevronDown size={16} className={clsx("text-foreground/60 transition-transform duration-200", menuOpen && "rotate-180")} />
                    </button>

                    {menuOpen && (
                      <div role="menu" className="absolute right-0 mt-2 w-72 rounded-2xl border border-foreground/10 bg-surface-elevated text-foreground shadow-2xl shadow-black/20 overflow-hidden z-50">
                        <div className="p-4 bg-gradient-to-r from-emerald-500/10 to-teal-500/10 border-b border-foreground/10">
                          <div className="flex items-center gap-3">
                            <div className="h-12 w-12 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center">
                              <span className="text-lg font-bold text-white">
                                {(user?.fullname || user?.email || '?').slice(0, 1).toUpperCase()}
                              </span>
                            </div>
                            <div className="flex-1 min-w-0">
                              <p className="font-semibold text-foreground truncate">{user?.fullname || '—'}</p>
                              <p className="text-xs text-foreground/50 truncate">{user?.email}</p>
                            </div>
                          </div>
                        </div>

                        <div className="p-2">
                          <Link
                            to="/settings/security"
                            className="flex items-center gap-3 rounded-xl px-3 py-2.5 hover:bg-foreground/5 text-foreground/70 hover:text-foreground transition-all"
                            onClick={() => setMenuOpen(false)}
                          >
                            <Settings size={18} />
                            <span>Account Settings</span>
                          </Link>
                        </div>

                        <div className="p-2 border-t border-foreground/10">
                          <button
                            className="w-full flex items-center gap-3 rounded-xl px-3 py-2.5 hover:bg-red-500/10 text-red-400 hover:text-red-300 transition-all"
                            onClick={onLogout}
                          >
                            <LogOut size={18} />
                            <span>Sign Out</span>
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </header>

          <main className="flex-1 px-4 md:px-6 pb-4 md:pb-6">
            <Outlet />
          </main>

          {/* ========== FOOTER ========== */}
          <footer className="mt-auto px-4 md:px-6 py-4 border-t border-foreground/5">
            <div className="flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-foreground/40">
              <div className="flex items-center gap-2">
                <span>&copy; {new Date().getFullYear()} SaaS Starter</span>
                <span className="hidden sm:inline">&bull;</span>
                <span className="hidden sm:inline">All rights reserved</span>
              </div>
              <div className="flex items-center gap-4">
                <a href="/privacy" className="hover:text-foreground/60 transition-colors">Privacy</a>
                <a href="/terms" className="hover:text-foreground/60 transition-colors">Terms</a>
                <span className="px-2 py-0.5 rounded bg-foreground/5 text-foreground/30">v1.0.0</span>
              </div>
            </div>
          </footer>
        </div>
      </div>

      {/* AI Chat Assistant — Org Admin only */}
      {isOrgAdmin && (
        <Suspense fallback={null}>
          <ChatWidget />
        </Suspense>
      )}
    </IdleTimeoutProvider>
  )
}
