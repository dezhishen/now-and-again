import { ref, watchEffect } from 'vue'

const isDark = ref(localStorage.getItem('na_dark') === 'true')

watchEffect(() => {
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('na_dark', String(isDark.value))
})

export function useDark() {
  return {
    isDark,
    toggle: () => { isDark.value = !isDark.value },
  }
}
