/**
 * Centralized UTC ↔ local timezone conversion.
 *
 * Backend stores and expects all times in UTC.
 * These helpers convert between local wall-clock strings ("HH:MM", "YYYY-MM-DD")
 * and their UTC equivalents at the API boundary.
 */

/** Convert local "HH:MM" → UTC "HH:MM". */
export function localTimeToUTC(localTime: string): string {
  const [h, m] = localTime.split(':').map(Number)
  const now = new Date()
  const d = new Date(now.getFullYear(), now.getMonth(), now.getDate(), h, m)
  return String(d.getUTCHours()).padStart(2, '0') + ':' + String(d.getUTCMinutes()).padStart(2, '0')
}

/** Convert UTC "HH:MM" → local "HH:MM". */
export function utcTimeToLocal(utcTime: string): string {
  const [h, m] = utcTime.split(':').map(Number)
  const now = new Date()
  const d = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate(), h, m))
  return String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
}

/** Convert local "YYYY-MM-DD" + "HH:MM" → UTC {date, time}. */
export function localDateTimeToUTC(localDate: string, localTime: string): { date: string; time: string } {
  const d = new Date(localDate + 'T' + localTime)
  const date = d.toISOString().slice(0, 10)
  const time = String(d.getUTCHours()).padStart(2, '0') + ':' + String(d.getUTCMinutes()).padStart(2, '0')
  return { date, time }
}

/** Convert UTC "YYYY-MM-DD" + "HH:MM" → local {date, time}. */
export function utcDateTimeToLocal(utcDate: string, utcTime: string): { date: string; time: string } {
  const d = new Date(utcDate + 'T' + utcTime + 'Z')
  const date = String(d.getFullYear()) + '-' +
    String(d.getMonth() + 1).padStart(2, '0') + '-' +
    String(d.getDate()).padStart(2, '0')
  const time = String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
  return { date, time }
}

// ── Deep conversion (used by API interceptor) ──────────────────

/**
 * Recursively walk an object/array and convert known time fields
 * in `schedule_data` from local → UTC before sending to backend.
 */
export function requestToUTC(obj: unknown): unknown {
  if (obj === null || obj === undefined) return obj
  if (Array.isArray(obj)) return obj.map(requestToUTC)
  if (typeof obj !== 'object') return obj

  const result: Record<string, unknown> = {}
  for (const key of Object.keys(obj as Record<string, unknown>)) {
    const val = (obj as Record<string, unknown>)[key]
    if (key === 'schedule_data' && val && typeof val === 'object' && !Array.isArray(val)) {
      const sd = { ...(val as Record<string, unknown>) }
      // Convert "HH:MM" time field
      if (typeof sd.time === 'string' && /^\d{2}:\d{2}$/.test(sd.time)) {
        sd.time = localTimeToUTC(sd.time)
      }
      // Convert date+time for one-shot tasks
      if (typeof sd.date === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(sd.date) &&
          typeof sd.time === 'string' && /^\d{2}:\d{2}$/.test(sd.time)) {
        const utc = localDateTimeToUTC(sd.date, sd.time as string)
        sd.date = utc.date
        sd.time = utc.time
      }
      result[key] = sd
    } else {
      result[key] = requestToUTC(val)
    }
  }
  return result
}

/**
 * Recursively walk API response data and convert known time fields
 * in `schedule_data` from UTC → local before returning to the UI.
 */
export function responseToLocal(obj: unknown): unknown {
  if (obj === null || obj === undefined) return obj
  if (Array.isArray(obj)) return obj.map(responseToLocal)
  if (typeof obj !== 'object') return obj

  const result: Record<string, unknown> = {}
  for (const key of Object.keys(obj as Record<string, unknown>)) {
    const val = (obj as Record<string, unknown>)[key]
    if (key === 'schedule_data' && val && typeof val === 'object' && !Array.isArray(val)) {
      const sd = { ...(val as Record<string, unknown>) }
      // Convert time field
      if (typeof sd.time === 'string' && /^\d{2}:\d{2}$/.test(sd.time)) {
        sd.time = utcTimeToLocal(sd.time)
      }
      // Convert date+time for one-shot tasks
      if (typeof sd.date === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(sd.date) &&
          typeof sd.time === 'string' && /^\d{2}:\d{2}$/.test(sd.time)) {
        const local = utcDateTimeToLocal(sd.date, sd.time as string)
        sd.date = local.date
        sd.time = local.time
      }
      result[key] = sd
    } else {
      result[key] = responseToLocal(val)
    }
  }
  return result
}
