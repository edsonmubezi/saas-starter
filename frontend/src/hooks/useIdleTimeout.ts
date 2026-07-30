import { useState, useEffect, useCallback, useRef } from 'react'

interface UseIdleTimeoutOptions {
  timeout: number // in milliseconds
  onIdle: () => void
  onActive?: () => void
  events?: string[]
}

const ACTIVITY_EVENTS = [
  'mousemove',
  'mousedown',
  'keydown',
  'touchstart',
  'scroll',
  'wheel',
]

// Throttle interval — activity is registered at most once per second
const THROTTLE_MS = 1000

export function useIdleTimeout({
  timeout,
  onIdle,
  onActive,
  events = ACTIVITY_EVENTS,
}: UseIdleTimeoutOptions) {
  const [isIdle, setIsIdle] = useState(false)
  const [remainingTime, setRemainingTime] = useState(timeout)

  const lastActivityRef = useRef<number>(Date.now())
  const isIdleRef = useRef(false)
  const lastThrottleRef = useRef<number>(0)

  // Keep callbacks in refs to avoid dependency churn
  const onIdleRef = useRef(onIdle)
  const onActiveRef = useRef(onActive)
  const timeoutRef = useRef(timeout)

  useEffect(() => { onIdleRef.current = onIdle }, [onIdle])
  useEffect(() => { onActiveRef.current = onActive }, [onActive])
  useEffect(() => { timeoutRef.current = timeout }, [timeout])

  const triggerLock = useCallback(() => {
    if (isIdleRef.current) return // already locked
    isIdleRef.current = true
    setIsIdle(true)
    setRemainingTime(0)
    onIdleRef.current()
  }, [])

  // Reset timer — called externally (e.g. on unlock)
  const resetTimer = useCallback(() => {
    lastActivityRef.current = Date.now()
    if (isIdleRef.current) {
      isIdleRef.current = false
      setIsIdle(false)
      onActiveRef.current?.()
    }
    setRemainingTime(timeoutRef.current)
  }, [])

  // Activity handler — just stamps the time, throttled to 1/sec
  const handleActivity = useCallback(() => {
    if (document.hidden) return
    if (isIdleRef.current) return

    const now = Date.now()
    if (now - lastThrottleRef.current < THROTTLE_MS) return
    lastThrottleRef.current = now
    lastActivityRef.current = now
  }, [])

  // Tab visibility — check elapsed on return
  const handleVisibility = useCallback(() => {
    if (document.hidden) return // going to background — polling will stop updating but that's fine

    // Tab becoming visible — check if we exceeded timeout while away
    const elapsed = Date.now() - lastActivityRef.current
    if (elapsed >= timeoutRef.current) {
      triggerLock()
    }
    // If not, the polling interval will pick it up
  }, [triggerLock])

  // Single polling interval — does all the work
  useEffect(() => {
    const id = setInterval(() => {
      if (isIdleRef.current) return // already locked

      const elapsed = Date.now() - lastActivityRef.current
      const remaining = Math.max(0, timeoutRef.current - elapsed)
      setRemainingTime(remaining)

      if (remaining <= 0) {
        triggerLock()
      }
    }, 1000)

    return () => clearInterval(id)
  }, [triggerLock])

  // Wire up activity events + visibility
  useEffect(() => {
    events.forEach((event) => {
      window.addEventListener(event, handleActivity, { passive: true })
    })
    document.addEventListener('visibilitychange', handleVisibility)

    return () => {
      events.forEach((event) => {
        window.removeEventListener(event, handleActivity)
      })
      document.removeEventListener('visibilitychange', handleVisibility)
    }
  }, [events, handleActivity, handleVisibility])

  return {
    isIdle,
    remainingTime,
    resetTimer,
  }
}

// Format remaining time as MM:SS
export function formatRemainingTime(ms: number): string {
  const totalSeconds = Math.ceil(ms / 1000)
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
}
