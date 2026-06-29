import type { ApiRequestError, ErrorCode, FieldError } from '@/types'
import type { I18nKey } from '@/i18n'

// ── ErrorCode → handler registry ─────────────────────────────────
// Each handler receives the full error + i18n translate function, returns a display message.

export type ErrorMessageHandler = (error: ApiRequestError, t: (key: I18nKey) => string) => string

function fieldsWithLabels(details: FieldError[], t: (key: I18nKey) => string): string {
  return details.map(d => {
    const labelKey = FIELD_LABEL_KEYS[d.field]
    const label = labelKey ? t(labelKey) : d.field
    return label + ': ' + d.message
  }).join('; ')
}

export const ERROR_HANDLERS: Record<ErrorCode, ErrorMessageHandler> = {
  BAD_REQUEST: (e) => e.summary,
  VALIDATION_ERROR: (e, t) => {
    if (!e.details?.length) return e.summary
    return fieldsWithLabels(e.details, t)
  },
  UNAUTHORIZED: (_e, t) => t('error.summary.unauthorized'),
  FORBIDDEN: (_e, t) => t('error.summary.forbidden'),
  NOT_FOUND: (_e, t) => t('error.summary.notFound'),
  CONFLICT: (_e, t) => t('error.summary.conflict'),
  INTERNAL_ERROR: (_e, t) => t('error.summary.internal'),
}

// ── Field name → i18n label key ──────────────────────────────────

const FIELD_LABEL_KEYS: Record<string, I18nKey> = {
  name: 'taskForm.taskName',
  displayName: 'register.displayName',
  username: 'login.username',
  password: 'login.password',
  email: 'register.email',
  phone: 'profile.phone',
  inviteCode: 'home.inviteCodePrefix',
  description: 'groups.desc',
  scheduleType: 'taskForm.schedule',
  kind: 'taskForm.schedule',
  title: 'taskForm.taskName',
}

/** Map a raw field name to a human-readable i18n label. */
export function fieldLabelKey(field: string): I18nKey | undefined {
  return FIELD_LABEL_KEYS[field]
}

/** Translate a FieldError into a display string. */
export function translateFieldError(f: FieldError, t: (key: I18nKey) => string): string {
  const labelKey = FIELD_LABEL_KEYS[f.field]
  const label = labelKey ? t(labelKey) : f.field
  return `${label}: ${f.message}`
}
