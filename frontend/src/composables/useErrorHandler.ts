import { computed, ref, shallowRef } from 'vue'
import { useI18n } from '@/i18n'
import type { I18nKey } from '@/i18n'
import { ApiRequestError } from '@/types'
import type { ErrorCode, FieldError } from '@/types'

// ── Error code → i18n summary key ────────────────────────────────

const ERROR_SUMMARY_KEYS: Record<ErrorCode, I18nKey> = {
  BAD_REQUEST: 'error.summary.badRequest',
  VALIDATION_ERROR: 'error.summary.validation',
  UNAUTHORIZED: 'error.summary.unauthorized',
  FORBIDDEN: 'error.summary.forbidden',
  NOT_FOUND: 'error.summary.notFound',
  CONFLICT: 'error.summary.conflict',
  INTERNAL_ERROR: 'error.summary.internal',
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
  kind: 'taskForm.schedule', // fallback
  title: 'taskForm.taskName',
}

export function useErrorHandler() {
  const { t } = useI18n()

  /** Current error to display (shallow to avoid deep reactivity on FieldError[]). */
  const error = shallowRef<ApiRequestError | null>(null)
  const expanded = ref(false)

  const summary = computed(() => {
    if (!error.value) return ''
    // Prefer backend summary if available
    if (error.value.summary && error.value.code !== 'INTERNAL_ERROR') {
      return error.value.summary
    }
    // Fallback to i18n
    const key = ERROR_SUMMARY_KEYS[error.value.code]
    return key ? t(key) : error.value.message
  })

  const detailItems = computed(() => {
    if (!error.value || !error.value.details?.length) return []
    const expanded = error.value.details.length <= 3
    return error.value.details.map((f: FieldError) => ({
      field: fieldLabel(f.field),
      message: f.message,
      expanded,
    }))
  })

  function setError(err: unknown) {
    if (err instanceof ApiRequestError) {
      error.value = err
      expanded.value = false
    } else if (err instanceof Error) {
      error.value = new ApiRequestError({ code: 'INTERNAL_ERROR', summary: err.message })
      expanded.value = false
    } else {
      error.value = null
    }
  }

  function clearError() {
    error.value = null
    expanded.value = false
  }

  function toggle() {
    expanded.value = !expanded.value
  }

  return { error, expanded, summary, detailItems, setError, clearError, toggle }
}

/** Map a raw field name to a human-readable i18n label. */
export function fieldLabel(field: string): string {
  const key = FIELD_LABEL_KEYS[field]
  if (key) {
    // We can't call useI18n() outside setup, so the caller passes t()
    return field // caller will translate
  }
  return field
}

/** Translate a FieldError into a display string using the i18n t function. */
export function translateFieldError(f: FieldError, t: (key: I18nKey) => string): string {
  const labelKey = FIELD_LABEL_KEYS[f.field]
  const label = labelKey ? t(labelKey) : f.field
  return `${label}: ${f.message}`
}
