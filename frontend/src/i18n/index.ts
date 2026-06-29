import { createI18n, useI18n as _useI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN'
import en from './locales/en'

/** Message schema — the Chinese locale is the canonical source of truth for keys. */
export type MessageSchema = typeof zhCN

/** Supported locale codes. */
export type SupportedLocale = 'zh-CN' | 'en'

// ─── Compile-time key-path extraction ───────────────────────────────────────

/**
 * Recursively build a union of all dot-separated paths in a nested object.
 *
 * Given { a: { b: string, c: string }, d: string },
 * yields `'a.b' | 'a.c' | 'd'`.
 */
type PathImpl<T, Key extends keyof T> =
  Key extends string
    ? T[Key] extends Record<string, any>
      ? Key | `${Key}.${PathImpl<T[Key], keyof T[Key]>}`
      : Key
    : never

/** All valid translation key paths (dot-separated). */
export type I18nKey = PathImpl<MessageSchema, keyof MessageSchema>

// ─── I18n instance ──────────────────────────────────────────────────────────

const saved = (localStorage.getItem('na_lang') || 'zh-CN') as SupportedLocale

const i18n = createI18n({
  legacy: false,
  locale: saved,
  fallbackLocale: 'zh-CN',
  messages: { 'zh-CN': zhCN, en },
})

// ─── Typed useI18n ──────────────────────────────────────────────────────────

/**
 * Typed replacement for vue-i18n's `useI18n`.
 *
 * The returned `t()` ONLY accepts valid translation key paths.
 * Invalid keys cause a TypeScript compile-time error.
 *
 * Use `td()` for keys computed at runtime (e.g. from task kind registries).
 *
 * ```ts
 * const { t, td } = useI18n()
 * t('login.heading')        // ✅ OK — known literal key
 * t('apiKey.scope.family')  // ✅ OK — 3-level nested key
 * // An invalid literal key would cause a compile error:
 * //   ❌ TS2345: Argument ... is not assignable to I18nKey
 * td(someDynamicKey)         // ✅ OK for runtime-computed keys
 * ```
 */
export function useI18n() {
  const composer = _useI18n()
  const rawT = composer.t
  return {
    ...composer,
    /** Type-safe translate: only accepts known i18n keys. */
    t: rawT as (key: I18nKey) => string,
    /** Dynamic translate: accepts any string. Use for runtime-computed keys. */
    td: rawT as (key: string) => string,
  } as Omit<typeof composer, 't'> & {
    t: (key: I18nKey) => string
    td: (key: string) => string
  }
}

export default i18n
