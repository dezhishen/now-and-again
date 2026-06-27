import { ref } from 'vue'

/**
 * Composable for automatic loading state management.
 *
 * Usage:
 *   const { loading, withLoading } = useLoading()
 *   await withLoading(() => api.get('/data'))
 *
 * The loading ref is true while any wrapped function is in flight.
 * Multiple concurrent calls are supported (loading stays true until all complete).
 */
export function useLoading() {
  const loading = ref(false)
  let pending = 0

  async function withLoading<T>(fn: () => Promise<T>): Promise<T> {
    pending++
    loading.value = true
    try {
      return await fn()
    } finally {
      pending--
      if (pending <= 0) {
        pending = 0
        loading.value = false
      }
    }
  }

  return { loading, withLoading }
}
