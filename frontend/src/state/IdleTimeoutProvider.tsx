import React, { createContext, useContext, useState, useCallback, useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from './AuthContext'
import { useIdleTimeout, formatRemainingTime } from '../hooks/useIdleTimeout'
import LockScreen from '../ui/LockScreen'
import { logout } from '../utils/api'
import { clearTokens, setSessionLocked, isSessionLocked, clearSessionLocked } from './tokenStore'
import { request, unwrap } from '../utils/common'
import toast from 'react-hot-toast'

// Default timeout while loading config from backend (15 minutes)
const DEFAULT_TIMEOUT_MS = 15 * 60 * 1000

// Warning threshold (show warning 1 minute before lock)
const WARNING_THRESHOLD = 60 * 1000

interface IdleTimeoutContextType {
  isLocked: boolean
  remainingTime: number
  lock: () => void
  unlock: () => void
}

const IdleTimeoutContext = createContext<IdleTimeoutContextType | undefined>(undefined)

interface IdleTimeoutProviderProps {
  children: React.ReactNode
  enabled?: boolean
}

export function IdleTimeoutProvider({
  children,
  enabled = true,
}: IdleTimeoutProviderProps) {
  const { user, refresh } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  // Initialize from localStorage to persist lock state across refresh
  const [isLocked, setIsLocked] = useState(() => isSessionLocked())
  const [hasShownWarning, setHasShownWarning] = useState(false)
  const [timeoutMs, setTimeoutMs] = useState(DEFAULT_TIMEOUT_MS)

  // Don't enable idle timeout on login page or when user is not authenticated
  const shouldEnableTimeout = enabled && !!user && !location.pathname.includes('/login')

  // Fetch session config from backend
  useEffect(() => {
    if (!shouldEnableTimeout) return

    let cancelled = false

    async function fetchConfig() {
      try {
        const data = await unwrap<{ idle_timeout_minutes: number }>(
          request('/auth/session-config')
        )
        if (!cancelled && data?.idle_timeout_minutes > 0) {
          setTimeoutMs(data.idle_timeout_minutes * 60 * 1000)
        }
      } catch {
        // Use default on error
      }
    }

    fetchConfig()

    // Refetch on window focus (in case admin changed the setting)
    function onFocus() { fetchConfig() }
    window.addEventListener('focus', onFocus)

    return () => {
      cancelled = true
      window.removeEventListener('focus', onFocus)
    }
  }, [shouldEnableTimeout])

  const handleIdle = useCallback(() => {
    if (shouldEnableTimeout) {
      setIsLocked(true)
      setSessionLocked(true)
      toast('Session locked due to inactivity', {
        icon: '🔒',
        duration: 3000,
      })
    }
  }, [shouldEnableTimeout])

  const handleActive = useCallback(() => {
    setHasShownWarning(false)
  }, [])

  const { remainingTime, resetTimer } = useIdleTimeout({
    timeout: timeoutMs,
    onIdle: handleIdle,
    onActive: handleActive,
  })

  // Show warning when approaching timeout
  useEffect(() => {
    if (
      shouldEnableTimeout &&
      !isLocked &&
      remainingTime <= WARNING_THRESHOLD &&
      remainingTime > 0 &&
      !hasShownWarning
    ) {
      setHasShownWarning(true)
      toast(
        `Your session will lock in ${formatRemainingTime(remainingTime)} due to inactivity`,
        {
          icon: '⚠️',
          duration: 5000,
        }
      )
    }
  }, [remainingTime, shouldEnableTimeout, isLocked, hasShownWarning])

  // Sync lock state across tabs/windows
  useEffect(() => {
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === 'session_locked') {
        const newLockState = e.newValue === 'true'
        setIsLocked(newLockState)
        if (!newLockState) {
          // Another tab unlocked, reset our timer too
          resetTimer()
          setHasShownWarning(false)
        }
      }
    }
    window.addEventListener('storage', handleStorageChange)
    return () => window.removeEventListener('storage', handleStorageChange)
  }, [resetTimer])

  const lock = useCallback(() => {
    setIsLocked(true)
    setSessionLocked(true)
  }, [])

  const unlock = useCallback(() => {
    setIsLocked(false)
    clearSessionLocked()
    resetTimer()
    setHasShownWarning(false)
  }, [resetTimer])

  const handleLogout = useCallback(async () => {
    try {
      await logout()
    } catch {
      // Ignore logout errors
    } finally {
      clearTokens()
      clearSessionLocked()
      await refresh()
      setIsLocked(false)
      navigate('/login', { replace: true })
    }
  }, [navigate, refresh])

  const value: IdleTimeoutContextType = {
    isLocked,
    remainingTime,
    lock,
    unlock,
  }

  // Show lock screen when locked
  if (isLocked && shouldEnableTimeout) {
    return (
      <IdleTimeoutContext.Provider value={value}>
        <LockScreen onUnlock={unlock} onLogout={handleLogout} />
      </IdleTimeoutContext.Provider>
    )
  }

  return (
    <IdleTimeoutContext.Provider value={value}>
      {children}
    </IdleTimeoutContext.Provider>
  )
}

export function useIdleTimeoutContext() {
  const context = useContext(IdleTimeoutContext)
  if (!context) {
    throw new Error('useIdleTimeoutContext must be used within IdleTimeoutProvider')
  }
  return context
}
