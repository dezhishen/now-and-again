import type { I18nKey } from '@/i18n'

type TranslateFn = (key: string) => string

/** Format a YYYY-MM-DD date string using the locale format. */
function fmtDate(d: string, t: TranslateFn): string {
  if (!d) return ''
  const parts = d.split('-')
  if (parts.length < 2) return d
  const m = parseInt(parts[1], 10)
  const day = parseInt(parts[2], 10)
  return t('schedule.dateFormat').replace('{m}', String(m)).replace('{d}', String(day))
}

export interface ScheduleTypeDef {
  value: string
  labelKey: I18nKey
  /** Builds a one-line summary, from large to small time units. Receives i18n translate function. */
  summary: (sd: Record<string, any>, t: TranslateFn) => string
}

const registry = new Map<string, ScheduleTypeDef>()

const SCHEDULE_TYPES: ScheduleTypeDef[] = [
  {
    value: 'once', labelKey: 'schedule.once',
    summary: (sd, t) => `${t('schedule.once')}${t('schedule.separator')}${fmtDate(sd.date, t)} ${sd.time || ''}`,
  },
  {
    value: 'daily', labelKey: 'schedule.daily',
    summary: (sd, t) => `${t('schedule.daily')}${t('schedule.separator')}${sd.time || ''}`,
  },
  {
    value: 'weekly', labelKey: 'schedule.weekly',
    summary: (sd, t) => {
      const wd = t('schedule.weekdays').split(',')
      const days = (sd.days || []).map((n: number) => wd[n - 1] || n).join('、')
      return `${t('schedule.weekly')} ${days}${t('schedule.separator')}${sd.time}`
    },
  },
  {
    value: 'monthly', labelKey: 'schedule.monthly',
    summary: (sd, t) => {
      const days = (sd.days || []).join('、')
      return `${t('schedule.monthly')} ${days}${t('schedule.dayUnit')}${t('schedule.separator')}${sd.time}`
    },
  },
  {
    value: 'yearly', labelKey: 'schedule.yearly',
    summary: (sd, t) => {
      const months = (sd.days || []).join('、')
      return `${t('schedule.yearly')} ${months}${t('schedule.monthUnit')} ${sd.day || 1}${t('schedule.dayUnit')}${t('schedule.separator')}${sd.time}`
    },
  },
  {
    value: 'interval', labelKey: 'schedule.interval',
    summary: (sd, t) => `${t('schedule.interval')}${sd.days || 1}${t('schedule.dayUnitInterval')}${t('schedule.separator')}${sd.time}`,
  },
]

SCHEDULE_TYPES.forEach(st => registry.set(st.value, st))

/** All registered schedule types (for select dropdowns). */
export function getScheduleTypes(): ScheduleTypeDef[] {
  return SCHEDULE_TYPES
}

/**
 * Builds a one-line schedule summary.
 * @param st  schedule_type value
 * @param sd  schedule_data object
 * @param t   i18n translate function
 */
export function scheduleLabel(st: string, sd: Record<string, any>, t: TranslateFn): string {
  const def = registry.get(st)
  if (!def) return st
  return def.summary(sd, t)
}
