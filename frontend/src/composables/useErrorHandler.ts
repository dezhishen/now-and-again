import { ref, type Ref } from 'vue'
import { ApiRequestError } from '@/types'
import type { ErrorCode, FieldError } from '@/types'
import type { I18nKey } from '@/i18n'

// ── Display mode ──────────────────────────────────────────────────

/** How an error should be shown to the user. */
export type DisplayMode = 'inline' | 'toast' | 'dialog'

/**
 * Registry mapping ErrorCode → DisplayMode.
 * Plugins or app initialisation can call registerDisplayMode() to override.
 */
const displayModes: Partial<Record<ErrorCode, DisplayMode>> = {
  BAD_REQUEST:      'toast',
  VALIDATION_ERROR: 'toast',
  UNAUTHORIZED:     'toast',
  FORBIDDEN:        'toast',
  NOT_FOUND:        'toast',
  CONFLICT:         'toast',
  INTERNAL_ERROR:   'toast',
}

/** Register or override the display mode for a given error code. */
export function registerDisplayMode(code: ErrorCode, mode: DisplayMode): void {
  displayModes[code] = mode
}

/** Resolve the display mode for an error, falling back to 'inline'. */
export function getDisplayMode(code: ErrorCode): DisplayMode {
  return displayModes[code] ?? 'inline'
}

// ── Severity ──────────────────────────────────────────────────────

export type Severity = 'info' | 'warning' | 'error' | 'success'

const severities: Partial<Record<ErrorCode, Severity>> = {
  BAD_REQUEST:      'warning',
  VALIDATION_ERROR: 'warning',
  UNAUTHORIZED:     'warning',
  FORBIDDEN:        'warning',
  NOT_FOUND:        'info',
  CONFLICT:         'warning',
  INTERNAL_ERROR:   'error',
}

/** Register or override the severity for a given error code. */
export function registerSeverity(code: ErrorCode, sev: Severity): void {
  severities[code] = sev
}

/** Resolve the severity for an error, falling back to 'warning'. */
export function getSeverity(code: ErrorCode): Severity {
  return severities[code] ?? 'warning'
}

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

// ── Public helpers ───────────────────────────────────────────────

/**
 * Convert an ApiRequestError to an i18n-translated display message.
 * Uses the ERROR_HANDLERS registry, falling back to raw summary if no handler is registered.
 */
export function formatApiError(error: ApiRequestError, t: (key: I18nKey) => string): string {
  const handler = ERROR_HANDLERS[error.code]
  return handler ? handler(error, t) : (error.summary || error.message)
}

/**
 * Composable that provides unified error handling for views.
 *
 * - `error` — reactive ref for binding to <ErrorDisplay :error="error" />
 * - `setError(e)` — capture any caught error into the ref, ErrorDisplay picks the display mode
 * - `clearError()` — dismiss the error
 *
 * All error display (toast / dialog / inline) is handled by <ErrorDisplay />.
 */
export function useErrorHandler() {
  const error: Ref<ApiRequestError | null> = ref(null)

  function toApiError(e: unknown): ApiRequestError {
    if (e instanceof ApiRequestError) return e as ApiRequestError
    if (e instanceof Error) return new ApiRequestError({ code: 'INTERNAL_ERROR', summary: e.message })
    return new ApiRequestError({ code: 'INTERNAL_ERROR', summary: String(e) })
  }

  function setError(e: unknown) {
    error.value = toApiError(e)
  }

  function clearError() {
    error.value = null
  }

  return { error, setError, clearError }
}
