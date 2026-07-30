import { createContext, useContext, useEffect, useRef, useState, useCallback } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { useAuth } from './AuthContext'
import { loadTokens } from './tokenStore'
import { tryRefresh } from '../utils/common'
import { API_BASE } from '../utils/common'

// ─── Types ───────────────────────────────────────────────────────────────────

export interface WSEvent {
  type: string
  payload?: unknown
  queryKeys?: string[]
  message?: string
  timestamp: number
}

export interface Notification {
  id: string
  type: string
  message: string
  timestamp: number
  read: boolean
}

interface WebSocketContextType {
  connected: boolean
  notifications: Notification[]
  unreadCount: number
  markAsRead: (id: string) => void
  markAllRead: () => void
  clearNotifications: () => void
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined)

const WS_CLOSE_AUTH_EXPIRED = 4001
const MAX_NOTIFICATIONS = 50
const MAX_RECONNECT_DELAY = 30000

function getWsUrl(): string {
  const base = API_BASE || window.location.origin
  return base.replace(/^http/, 'ws') + '/ws'
}

// ─── Provider ────────────────────────────────────────────────────────────────

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const [connected, setConnected] = useState(false)
  const [notifications, setNotifications] = useState<Notification[]>([])
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectAttemptRef = useRef(0)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout>>()
  const mountedRef = useRef(true)

  const unreadCount = notifications.filter(n => !n.read).length

  const markAsRead = useCallback((id: string) => {
    setNotifications(prev => prev.map(n => n.id === id ? { ...n, read: true } : n))
  }, [])

  const markAllRead = useCallback(() => {
    setNotifications(prev => prev.map(n => ({ ...n, read: true })))
  }, [])

  const clearNotifications = useCallback(() => {
    setNotifications([])
  }, [])

  const handleEvent = useCallback((event: WSEvent) => {
    // 1. Invalidate TanStack Query keys
    if (event.queryKeys?.length) {
      for (const key of event.queryKeys) {
        queryClient.invalidateQueries({ queryKey: [key] })
      }
    }

    // 2. Show toast
    if (event.message) {
      if (event.type.includes('reject')) {
        toast.error(event.message, { duration: 5000 })
      } else {
        toast.success(event.message, { duration: 4000 })
      }
    }

    // 3. Add to notification list
    if (event.message) {
      const notification: Notification = {
        id: `${event.timestamp}-${Math.random().toString(36).slice(2, 8)}`,
        type: event.type,
        message: event.message,
        timestamp: event.timestamp,
        read: false,
      }
      setNotifications(prev => [notification, ...prev].slice(0, MAX_NOTIFICATIONS))
    }
  }, [queryClient])

  const connect = useCallback(() => {
    const tokens = loadTokens()
    if (!tokens.access || !user) return

    // Clean up any existing connection
    if (wsRef.current) {
      wsRef.current.onclose = null
      wsRef.current.close()
      wsRef.current = null
    }

    const url = `${getWsUrl()}?token=${encodeURIComponent(tokens.access)}`
    const ws = new WebSocket(url)

    ws.onopen = () => {
      if (!mountedRef.current) return
      setConnected(true)
      reconnectAttemptRef.current = 0
    }

    ws.onmessage = (evt) => {
      try {
        const event: WSEvent = JSON.parse(evt.data)
        handleEvent(event)
      } catch {
        // Ignore malformed messages
      }
    }

    ws.onclose = async (evt) => {
      if (!mountedRef.current) return
      setConnected(false)
      wsRef.current = null

      if (evt.code === WS_CLOSE_AUTH_EXPIRED) {
        const ok = await tryRefresh()
        if (ok && mountedRef.current) {
          scheduleReconnect(0)
        }
        return
      }

      // Reconnect with exponential backoff
      if (user && mountedRef.current) {
        scheduleReconnect()
      }
    }

    ws.onerror = () => {
      // onclose fires after onerror — reconnect handled there
    }

    wsRef.current = ws
  }, [user, handleEvent])

  const scheduleReconnect = useCallback((overrideDelay?: number) => {
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current)
    }
    const delay = overrideDelay ?? Math.min(
      1000 * Math.pow(2, reconnectAttemptRef.current),
      MAX_RECONNECT_DELAY
    )
    reconnectAttemptRef.current += 1
    reconnectTimerRef.current = setTimeout(() => {
      if (mountedRef.current) connect()
    }, delay)
  }, [connect])

  useEffect(() => {
    mountedRef.current = true
    if (user) {
      connect()
    }
    return () => {
      mountedRef.current = false
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current)
      }
      if (wsRef.current) {
        wsRef.current.onclose = null
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [user, connect])

  return (
    <WebSocketContext.Provider value={{
      connected,
      notifications,
      unreadCount,
      markAsRead,
      markAllRead,
      clearNotifications,
    }}>
      {children}
    </WebSocketContext.Provider>
  )
}

// ─── Hook ────────────────────────────────────────────────────────────────────

export function useWebSocket() {
  const ctx = useContext(WebSocketContext)
  if (!ctx) throw new Error('useWebSocket must be used within WebSocketProvider')
  return ctx
}
