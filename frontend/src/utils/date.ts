// src/utils/date.ts
/**
 * Format a date-like input as "Month Year" (e.g., "September 2025").
 * Accepts Date, epoch (ms or sec), or common strings like:
 *  - "YYYY-MM"
 *  - "YYYYMM"
 *  - "YYYY-MM-DD" / "YYYY/MM/DD"
 */
export function formatMonthYear(
  input: string | number | Date,
  options?: { locale?: string | string[]; fallback?: string }
): string {
  const { locale, fallback } = options ?? {}
  const d = coerceToDate(input)
  if (!d) return fallback ?? String(input ?? '')
  return new Intl.DateTimeFormat(locale, { month: 'long', year: 'numeric' }).format(d)
}

function coerceToDate(input: string | number | Date): Date | null {
  if (input instanceof Date) return isNaN(input.getTime()) ? null : input

  if (typeof input === 'number') {
    // support seconds or milliseconds
    const ms = input > 1e12 ? input : input * 1000
    const d = new Date(ms)
    return isNaN(d.getTime()) ? null : d
  }

  if (typeof input === 'string') {
    const s = input.trim()
    let d: Date | null = null

    if (/^\d{4}-\d{2}$/.test(s)) {
      // "YYYY-MM"
      d = new Date(`${s}-01T00:00:00`)
    } else if (/^\d{4}\/\d{2}$/.test(s)) {
      // "YYYY/MM"
      d = new Date(`${s}/01 00:00:00`)
    } else if (/^\d{6}$/.test(s)) {
      // "YYYYMM"
      d = new Date(`${s.slice(0, 4)}-${s.slice(4, 6)}-01T00:00:00`)
    } else if (/^\d{4}[-/]\d{2}[-/]\d{2}/.test(s)) {
      // "YYYY-MM-DD" or "YYYY/MM/DD"
      d = new Date(s)
    } else {
      const maybe = new Date(s)
      d = isNaN(maybe.getTime()) ? null : maybe
    }
    return d && !isNaN(d.getTime()) ? d : null
  }

  return null
}


// src/utils/date.ts

/**
 * Format a date-like input as "DD.Month.Year" (e.g., "13.September.2025").
 * Accepts Date, epoch (ms or sec), or common strings.
 */
export function formatDayMonthYear(
  input: string | number | Date,
  options?: { locale?: string | string[]; fallback?: string }
): string {
  const { locale, fallback } = options ?? {}
  const d = coerceToDate(input)
  if (!d) return fallback ?? String(input ?? '')

  const day = d.getDate()
  const month = new Intl.DateTimeFormat(locale, { month: 'long' }).format(d)
  const year = d.getFullYear()

  return `${day}.${month}.${year}`
}


export function formatDayMonthYearShort(
  input: string | number | Date,
  options?: { locale?: string | string[]; fallback?: string }
): string {
  const { locale, fallback } = options ?? {}
  const d = coerceToDate(input)
  if (!d) return fallback ?? String(input ?? '')

  const day = d.getDate()
  const month = new Intl.DateTimeFormat(locale, { month: 'short' }).format(d)
  const year = d.getFullYear()

  return `${day}.${month}.${year}`
}

/**
 * Format a date-like input as "dd-mm-yyyy" (e.g., "13-09-2025").
 * Accepts Date, epoch (ms or sec), or common strings like "2025-09-13 21:26:00".
 */
export function formatDateDDMMYYYY(
  input: string | number | Date,
  options?: { fallback?: string }
): string {
  const { fallback } = options ?? {}
  const d = coerceToDate(input)
  if (!d) return fallback ?? String(input ?? '')

  const day = d.getDate().toString().padStart(2, '0')
  const month = (d.getMonth() + 1).toString().padStart(2, '0')
  const year = d.getFullYear()

  return `${day}-${month}-${year}`
}

/**
 * Format a date-like input to show time as "HH:MM" (e.g., "21:26").
 * Accepts Date, epoch (ms or sec), or common strings like "2025-09-13 21:26:00".
 */
export function formatTime(
  input: string | number | Date,
  options?: { fallback?: string }
): string {
  const { fallback } = options ?? {}
  const d = coerceToDate(input)
  if (!d) return fallback ?? String(input ?? '')

  const hours = d.getHours().toString().padStart(2, '0')
  const minutes = d.getMinutes().toString().padStart(2, '0')

  return `${hours}:${minutes}`
}

/**
 * Format a date-like input as "dd-mm-yyyy HH:MM" (e.g., "13-09-2025 21:26").
 * Accepts Date, epoch (ms or sec), or common strings like "2025-09-13 21:26:00".
 */
export function formatDateTimeShort(
  input: string | number | Date,
  options?: { fallback?: string }
): string {
  const { fallback } = options ?? {}
  const d = coerceToDate(input)
  if (!d) return fallback ?? String(input ?? '')

  const dateStr = formatDateDDMMYYYY(d)
  const timeStr = formatTime(d)

  return `${dateStr} ${timeStr}`
}
