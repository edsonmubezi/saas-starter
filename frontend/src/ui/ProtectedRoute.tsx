
import React, { useState, useCallback, useEffect } from 'react'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../state/AuthContext'
import { isSessionLocked, clearSessionLocked, clearTokens, hasValidTokens } from '../state/tokenStore'
import { logout } from '../utils/api'
import LockScreen from './LockScreen'

export default function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading, refresh } = useAuth()
  const loc = useLocation()
  const nav = useNavigate()

  // Check session lock state from localStorage BEFORE anything else.
  // This ensures new tabs immediately see the lock screen instead of
  // trying to load the app (which may fail if the JWT expired).
  const [locked, setLocked] = useState(() => isSessionLocked())

  // Sync lock state across tabs via storage events
  useEffect(() => {
    const onStorage = (e: StorageEvent) => {
      if (e.key === 'session_locked') {
        setLocked(e.newValue === 'true')
      }
    }
    window.addEventListener('storage', onStorage)
    return () => window.removeEventListener('storage', onStorage)
  }, [])

  const handleUnlock = useCallback(() => {
    setLocked(false)
    clearSessionLocked()
  }, [])

  const handleLogout = useCallback(async () => {
    try { await logout() } catch { /* ignore */ }
    clearTokens()
    clearSessionLocked()
    setLocked(false)
    await refresh()
    nav('/login', { replace: true })
  }, [nav, refresh])

  // If session is locked AND we still have tokens, show lock screen right away
  if (locked && hasValidTokens()) {
    return <LockScreen onUnlock={handleUnlock} onLogout={handleLogout} />
  }

  // If session was locked but tokens expired, clear lock and force re-login
  if (locked && !hasValidTokens()) {
    clearSessionLocked()
    clearTokens()
    return <Navigate to="/login" replace />
  }

  if (loading) return <div className="min-h-screen grid place-items-center bg-surface-primary text-foreground/80">Loading…</div>
  if (!user) { const next = encodeURIComponent(loc.pathname + loc.search); return <Navigate to={`/login?next=${next || '%2F'}`} replace /> }
  return <>{children}</>
}
