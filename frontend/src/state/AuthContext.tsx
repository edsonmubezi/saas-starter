import React, { createContext, useContext, useEffect, useState, useMemo } from 'react'
import { getMe } from '../utils/api'
import type { Me } from '../utils/types'
import { hasValidTokens, clearTokens, isAccountLocked, clearAccountLocked, isTokenExpired, loadTokens } from './tokenStore'
import { getApiPrefix, getUserType, API_PREFIX, type ApiPrefixType } from '../utils/apiPrefix'
import { resetAuthExpired } from '../utils/common'

type AuthContextType = {
  user: Me | null
  loading: boolean
  refresh: () => Promise<Me | null>
  /** The API prefix based on user permissions (/api/admin or /api/org) */
  apiPrefix: ApiPrefixType
  /** The user type based on permissions ('admin' or 'org') */
  userType: 'admin' | 'org'
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<Me | null>(null)
  const [loading, setLoading] = useState(true)

  async function refresh(): Promise<Me | null> {
    try {
      // Check token validity before making API call
      if (!hasValidTokens()) {
        clearTokens()
        setUser(null)
        setLoading(false)
        return null
      }

      const me = await getMe()
      resetAuthExpired() // clear the expired flag after successful auth
      setUser(me)
      return me
    } catch {
      clearTokens()
      setUser(null)
      return null
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    let mounted = true

    const run = async () => {
      try {
        // Check if account was previously marked as locked
        if (isAccountLocked()) {
          clearTokens()
          clearAccountLocked()
          if (mounted) {
            setUser(null)
            setLoading(false)
          }
          // Redirect to login if not already there
          if (window.location.pathname !== '/login') {
            window.location.href = '/login'
          }
          return
        }

        // Check token validity before making API call
        if (!hasValidTokens()) {
          clearTokens()
          if (mounted) {
            setUser(null)
            setLoading(false)
          }
          return
        }

        const me = await getMe()
        resetAuthExpired()
        if (mounted) setUser(me)
      } catch {
        clearTokens()
        if (mounted) setUser(null)
      } finally {
        if (mounted) setLoading(false)
      }
    }
    run()

    // Keep auth state in sync across tabs/windows
    const onStorage = (e: StorageEvent) => {
      if (e.key === 'auth_tokens') {
        // tokens changed in another tab — re-check current user
        if (!hasValidTokens()) {
          // Tokens cleared or expired in another tab
          if (mounted) { setUser(null); setLoading(false) }
        } else {
          setLoading(true)
          refresh()
        }
      }
      // Handle account lock changes from other tabs
      if (e.key === 'account_locked' && e.newValue === 'true') {
        clearTokens()
        clearAccountLocked()
        if (mounted) { setUser(null); setLoading(false) }
      }
    }
    window.addEventListener('storage', onStorage)

    // Listen for auth:expired events from the request utility (instant, no full page reload)
    const onAuthExpired = () => {
      clearTokens()
      if (mounted) { setUser(null); setLoading(false) }
    }
    window.addEventListener('auth:expired', onAuthExpired)

    // Proactive token expiry check — detect expiry before API calls fail
    const expiryCheck = setInterval(() => {
      const tokens = loadTokens()
      if (tokens.access && isTokenExpired()) {
        clearTokens()
        if (mounted) { setUser(null); setLoading(false) }
      }
    }, 15_000) // check every 15 seconds

    return () => {
      mounted = false
      window.removeEventListener('storage', onStorage)
      window.removeEventListener('auth:expired', onAuthExpired)
      clearInterval(expiryCheck)
    }
  }, [])

  // Compute API prefix based on user permissions
  const apiPrefix = useMemo(() => {
    return getApiPrefix(user?.permissions ?? [])
  }, [user?.permissions])

  const userType = useMemo(() => {
    return getUserType(user?.permissions ?? [])
  }, [user?.permissions])

  return (
    <AuthContext.Provider value={{ user, loading, refresh, apiPrefix, userType }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
