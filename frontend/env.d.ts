/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module 'vue-i18n' {
  export function createI18n(options: Record<string, unknown>): {
    install: (app: unknown) => void
    global: { t: (key: string) => string; locale: { value: string } }
  }
  export function useI18n(): { t: (key: string) => string; locale: { value: string } }
}
