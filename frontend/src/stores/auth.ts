import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import type { User, Family, SystemStatus } from '@/types'
import { api } from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const families = ref<Family[]>([])
  const activeFamilyId = ref<string | null>(null)
  const initialized = ref<boolean | null>(null)
  const error = ref<string | null>(null)
  const sessionChecked = ref(false)

  const isLoggedIn = computed(() => !!user.value && api.hasValidToken())
  const needsSetup = computed(() => initialized.value === false)
  const isAdmin = computed(() => user.value?.roles?.includes('admin') ?? false)

  api.onExpired(() => {
    user.value = null
    const router = useRouter()
    router.push('/login')
  })

  async function checkInit() {
    try {
      const status = await api.get<SystemStatus>('/system/status')
      initialized.value = status.initialized
    } catch { initialized.value = false }
  }

  async function initSession() {
    if (sessionChecked.value) return
    sessionChecked.value = true
    const u = await api.initSession()
    if (u) user.value = u
  }

  async function setup(req: { username: string; email: string; password: string; display_name: string }) {
    error.value = null
    try {
      await api.post('/setup', req)
      initialized.value = true
      return true
    } catch (e: any) { error.value = e.message; return false }
  }

  async function register(req: { username: string; email: string; password: string; display_name: string }) {
    error.value = null
    try {
      await api.post('/auth/register', req)
      return true
    } catch (e: any) { error.value = e.message; return false }
  }

  async function login(username: string, password: string) {
    error.value = null
    try {
      const data = await api.post<{ access_token: string; expires_in: number; user: User }>(
        '/auth/login', { username, password }
      )
      api.setAccessToken(data.access_token)
      user.value = data.user
      return true
    } catch (e: any) { error.value = e.message; return false }
  }

  async function logout() {
    try { await api.post('/auth/logout') } catch { /* */ }
    api.setAccessToken(null)
    user.value = null
    families.value = []
    activeFamilyId.value = null
  }

  return {
    user, families, activeFamilyId, initialized, needsSetup, error, sessionChecked, isLoggedIn, isAdmin,
    checkInit, initSession, setup, register, login, logout,
  }
})
