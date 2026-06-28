import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import type { User, Family } from '@/types'
import { api } from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const families = ref<Family[]>([])
  const activeFamilyId = ref<string | null>(null)
  const sessionChecked = ref(false)

  const isLoggedIn = computed(() => api.hasValidToken())
  const isAdmin = computed(() => user.value?.roles?.includes('admin') ?? false)

  // ── Session expiry callback (registered once) ──────────────
  api.onExpired(() => {
    user.value = null
    useRouter().push('/login')
  })

  // ── silent token restore (called by router guard) ──────────
  async function initSession() {
    if (sessionChecked.value) return
    sessionChecked.value = true
    const u = await api.initSession()
    if (u) user.value = u
  }

  async function fetchUser() {
    if (user.value) return
    try {
      user.value = await api.get<User>('/users/me')
    } catch {
      // Token is valid but user doesn't exist (e.g. db-reset).
      // Clear the stale token so the guard redirects to login.
      api.setAccessToken(null)
      sessionChecked.value = false
    }
  }

  // ── register ───────────────────────────────────────────────

  async function register(req: {
    username: string; email: string; password: string; display_name: string
  }) {
    await api.post('/auth/register', req)
  }

  // ── login ──────────────────────────────────────────────────

  /** POST /auth/login, store token + user. Throws on failure. */
  async function login(username: string, password: string) {
    const data = await api.post<{ access_token: string; user: User }>(
      '/auth/login',
      { username, password },
    )
    api.setAccessToken(data.access_token)
    user.value = data.user
    sessionChecked.value = true
  }

  // ── logout ─────────────────────────────────────────────────

  /** Clear local state immediately, then invalidate server-side (best-effort). */
  async function logout() {
    api.setAccessToken(null)
    user.value = null
    families.value = []
    activeFamilyId.value = null
    sessionChecked.value = false
    try { await api.post('/auth/logout') } catch { /* best-effort */ }
  }

  return {
    user, families, activeFamilyId, sessionChecked,
    isLoggedIn, isAdmin,
    initSession, fetchUser, register, login, logout,
  }
})
