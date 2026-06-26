import { ref } from 'vue'

export type ToastLevel = 'success' | 'error' | 'warning' | 'info'

interface Toast {
  id: number
  message: string
  level: ToastLevel
}

const toasts = ref<Toast[]>([])
let nextId = 0

function add(level: ToastLevel, message: string) {
  const id = nextId++
  toasts.value.push({ id, message, level })
  setTimeout(() => remove(id), 4000)
}

function remove(id: number) {
  toasts.value = toasts.value.filter(t => t.id !== id)
}

export function useToast() {
  return {
    toasts,
    success: (msg: string) => add('success', msg),
    error: (msg: string) => add('error', msg),
    warning: (msg: string) => add('warning', msg),
    info: (msg: string) => add('info', msg),
    remove,
  }
}
