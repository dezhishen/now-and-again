import { ref, computed, watchEffect } from 'vue'

type Theme = 'light' | 'dark' | 'auto'

const STORAGE_KEY = 'na_theme'
const stored = localStorage.getItem(STORAGE_KEY) as Theme | null
const theme = ref<Theme>(stored || 'auto')
const isDark = ref(false)

function applyTheme() {
  let dark: boolean
  if (theme.value === 'auto') {
    dark = window.matchMedia('(prefers-color-scheme: dark)').matches
  } else {
    dark = theme.value === 'dark'
  }
  isDark.value = dark
  document.documentElement.classList.toggle('dark', dark)
}

applyTheme()
watchEffect(() => {
  localStorage.setItem(STORAGE_KEY, theme.value)
  applyTheme()
})

window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
  if (theme.value === 'auto') applyTheme()
})

export function useDark() {
  const label = computed(() => {
    switch (theme.value) {
      case 'light': return '☀️'
      case 'dark': return '🌙'
      case 'auto': return '💻'
    }
  })
  const title = computed(() => {
    switch (theme.value) {
      case 'light': return '浅色模式'
      case 'dark': return '深色模式'
      case 'auto': return '跟随系统'
    }
  })

  function cycle() {
    const order: Theme[] = ['light', 'dark', 'auto']
    const idx = order.indexOf(theme.value)
    theme.value = order[(idx + 1) % order.length]
  }

  return { isDark, theme, cycle, label, title }
}
