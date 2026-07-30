import { useState, useRef, useEffect } from 'react'
import { Bell, CheckCheck, Trash2, Wifi, WifiOff } from 'lucide-react'
import clsx from 'clsx'
import { useWebSocket, type Notification } from '../state/WebSocketContext'

function timeAgo(timestamp: number): string {
  const diff = Date.now() - timestamp
  const seconds = Math.floor(diff / 1000)
  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  return `${Math.floor(hours / 24)}d ago`
}

function eventColor(type: string): string {
  if (type.startsWith('leave')) return 'bg-violet-500'
  if (type.startsWith('loan')) return 'bg-amber-500'
  if (type.startsWith('attendance')) return 'bg-emerald-500'
  if (type.startsWith('payroll')) return 'bg-blue-500'
  return 'bg-foreground/40'
}

export default function NotificationBell() {
  const { connected, notifications, unreadCount, markAsRead, markAllRead, clearNotifications } = useWebSocket()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen(o => !o)}
        className="relative p-2.5 rounded-xl bg-foreground/5 hover:bg-foreground/10 border border-foreground/10 hover:border-foreground/20 transition-all group"
        title="Notifications"
      >
        <Bell size={18} className="text-foreground/60 group-hover:text-foreground transition-colors" />
        {unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 min-w-[18px] h-[18px] flex items-center justify-center rounded-full bg-red-500 text-white text-[10px] font-bold px-1 leading-none">
            {unreadCount > 99 ? '99+' : unreadCount}
          </span>
        )}
        <span className={clsx(
          'absolute bottom-1.5 right-1.5 w-1.5 h-1.5 rounded-full',
          connected ? 'bg-emerald-400' : 'bg-red-400'
        )} />
      </button>

      {open && (
        <div className="absolute right-0 mt-2 w-80 max-h-[420px] rounded-2xl border border-foreground/10 bg-surface-elevated text-foreground shadow-2xl shadow-black/20 overflow-hidden z-50">
          {/* Header */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-foreground/[0.06] bg-foreground/[0.02]">
            <div className="flex items-center gap-2">
              <h3 className="text-sm font-semibold">Notifications</h3>
              <span className="flex items-center gap-1 text-[10px] text-foreground/40">
                {connected ? <Wifi size={10} /> : <WifiOff size={10} />}
                {connected ? 'Live' : 'Offline'}
              </span>
            </div>
            <div className="flex items-center gap-1">
              {unreadCount > 0 && (
                <button
                  onClick={markAllRead}
                  className="p-1.5 rounded-lg hover:bg-foreground/5 text-foreground/40 hover:text-foreground/60 transition-all"
                  title="Mark all read"
                >
                  <CheckCheck size={14} />
                </button>
              )}
              {notifications.length > 0 && (
                <button
                  onClick={clearNotifications}
                  className="p-1.5 rounded-lg hover:bg-foreground/5 text-foreground/40 hover:text-foreground/60 transition-all"
                  title="Clear all"
                >
                  <Trash2 size={14} />
                </button>
              )}
            </div>
          </div>

          {/* List */}
          <div className="overflow-y-auto max-h-[340px]">
            {notifications.length === 0 ? (
              <div className="py-10 text-center text-foreground/30 text-sm">
                No notifications yet
              </div>
            ) : (
              notifications.map((n: Notification) => (
                <div
                  key={n.id}
                  onClick={() => markAsRead(n.id)}
                  className={clsx(
                    'px-4 py-3 border-b border-foreground/[0.04] cursor-pointer hover:bg-foreground/[0.04] transition-all',
                    !n.read && 'bg-blue-500/[0.04]'
                  )}
                >
                  <div className="flex items-start gap-3">
                    <span className={clsx(
                      'mt-1.5 w-2 h-2 rounded-full shrink-0',
                      !n.read ? eventColor(n.type) : 'bg-transparent'
                    )} />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm text-foreground/80 leading-snug">{n.message}</p>
                      <p className="text-[10px] text-foreground/30 mt-1">{timeAgo(n.timestamp)}</p>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  )
}
