import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import type { User, Family } from '@/types'
import { api } from '@/api/client'

const FAMILY_KEY = 'na_active_family'

function loadFamilyId(): string | null {
  try { return localStorage.getItem(FAMILY_KEY) } catch { return null }
}
function saveFamilyId(id: string | null) {
  try {
    if (id) localStorage.setItem(FAMILY_KEY, id)
    else localStorage.removeItem(FAMILY_KEY)
  } catch { /* */ }
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const families = ref<Family[]>([])
  const activeFamilyId = ref<string | null>(loadFamilyId())
  const sessionChecked = ref(false)

  // Sync family ID to API header
  watch(activeFamilyId, (id) => {
    api.setFamilyId(id)
    saveFamilyId(id)
  }, { immediate: true })

  const isLoggedIn = computed(() => api.hasValidToken())
  const isAdmin = computed(() => user.value?.roles?.includes('admin') ?? false)

  // ── Session expiry callback (registered once) ──────────────
  api.onExpired(() => {
    user.value = null
    window.location.href = '/login'
  })

  // ── Family not found callback (registered once) ────────────
  api.onFamilyNotFound(() => {
    activeFamilyId.value = null
    window.location.href = '/families'
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

  // ── family ────────────────────────────────────────────────

  async function loadFamilies() {
    try { families.value = await api.get<Family[]>('/users/me/families') } catch { /* */ }
  }

  function switchFamily(id: string) {
    activeFamilyId.value = id
  }

  /**
   * Resolve the active family after login / session restore.
   *
   * 1. If localStorage holds a previously-used family ID:
   *    - Fetch that family via GET /api/families/:id
   *    - If archived or 404 → clear the saved ID, return null
   *    - Otherwise keep it as the active family
   * 2. If no saved ID: fall back to the user's default_family_id
   *    from /users/me
   *
   * Returns the resolved family ID, or null if none could be determined.
   * Caller should redirect to /families when null.
   */
  async function resolveFamily(): Promise<string | null> {
    // ── Case 1: localStorage has a saved family ID → validate it ──
    if (activeFamilyId.value) {
      try {
        const family = await api.get<Family>(`/families/${activeFamilyId.value}`)
        if (family.archived) {
          activeFamilyId.value = null
          return null
        }
        return activeFamilyId.value
      } catch {
        // 404 (family deleted) or network error → clear stale ID
        activeFamilyId.value = null
        return null
      }
    }

    // ── Case 2: no saved ID → use default from /me ──
    if (user.value?.default_family_id) {
      // Also validate the default family (it might be archived too)
      try {
        const family = await api.get<Family>(`/families/${user.value.default_family_id}`)
        if (family.archived) return null
        activeFamilyId.value = user.value.default_family_id
        return user.value.default_family_id
      } catch {
        return null
      }
    }

    return null
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
    loadFamilies, switchFamily, resolveFamily,
  }
})
