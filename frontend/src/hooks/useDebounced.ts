// =============================
// Shared files
// =============================

// File: src/hooks/useDebounced.ts
import * as React from 'react'
export function useDebounced<T>(value: T, delay = 400) {
  const [v, setV] = React.useState(value)
  React.useEffect(() => {
    const t = setTimeout(() => setV(value), delay)
    return () => clearTimeout(t)
  }, [value, delay])
  return v
}
