import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import i18n from './i18n'
import App from './App.vue'
import './styles/global.css'

const app = createApp(App)

// Global v-esc directive: closes the topmost dialog on Escape key.
// Supports nested dialogs — only the innermost one closes.
const escStack: Array<() => void> = []

function onEscKey(e: KeyboardEvent) {
  if (e.key === 'Escape' && escStack.length > 0) {
    escStack[escStack.length - 1]()
  }
}
document.addEventListener('keydown', onEscKey)

app.directive('esc', {
  mounted(_el, binding) {
    escStack.push(binding.value)
  },
  unmounted(_el, binding) {
    const idx = escStack.lastIndexOf(binding.value)
    if (idx >= 0) escStack.splice(idx, 1)
  },
})

app.use(createPinia())
app.use(router)
app.use(i18n)
app.mount('#app')
