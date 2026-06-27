/**
 * Centralized datetime formatting utilities.
 *
 * All backend timestamps are RFC 3339 UTC (e.g., "2026-06-27T09:00:00Z").
 * These helpers parse them correctly and format for the user's locale.
 */

const DEFAULT_LOCALE = 'zh-CN'

const FULL_OPTS: Intl.DateTimeFormatOptions = {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
}

const DATE_OPTS: Intl.DateTimeFormatOptions = {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
}

const TIME_OPTS: Intl.DateTimeFormatOptions = {
  hour: '2-digit',
  minute: '2-digit',
}

/** Format an RFC 3339 string as full date+time in user locale. */
export function fmtDateTime(iso: string | null | undefined, locale = DEFAULT_LOCALE): string {
  if (!iso) return '-'
  return new Date(iso).toLocaleString(locale, FULL_OPTS)
}

/** Format an RFC 3339 string as date only. */
export function fmtDate(iso: string | null | undefined, locale = DEFAULT_LOCALE): string {
  if (!iso) return '-'
  return new Date(iso).toLocaleDateString(locale, DATE_OPTS)
}

/** Format an RFC 3339 string as time only. */
export function fmtTime(iso: string | null | undefined, locale = DEFAULT_LOCALE): string {
  if (!iso) return '-'
  return new Date(iso).toLocaleTimeString(locale, TIME_OPTS)
}

/** Format as "MM-DD HH:mm → MM-DD HH:mm" range. */
export function fmtRange(start: string, end: string, locale = DEFAULT_LOCALE): string {
  const s = new Date(start)
  const e = new Date(end)
  const opts: Intl.DateTimeFormatOptions = { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }
  return s.toLocaleDateString(locale, opts) + ' → ' + e.toLocaleDateString(locale, opts)
}

/** Quick relative time (e.g., "3分钟前"). Input is RFC 3339 string. */
export function fmtRelative(iso: string | null | undefined): string {
  if (!iso) return '-'
  const now = Date.now()
  const then = new Date(iso).getTime()
  const diff = now - then
  const seconds = Math.floor(diff / 1000)
  if (seconds < 60) return '刚刚'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}天前`
  return fmtDate(iso)
}

/** Convert a local datetime-local input value to RFC 3339 UTC string for API. */
export function localToUTC(localStr: string): string {
  if (!localStr) return ''
  return new Date(localStr).toISOString()
}

/** Convert an RFC 3339 UTC string to datetime-local input value. */
export function utcToLocal(iso: string | null | undefined): string {
  if (!iso) return ''
  const d = new Date(iso)
  // Format as YYYY-MM-DDTHH:mm (local time)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}
