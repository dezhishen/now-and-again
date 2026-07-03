<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from '@/i18n'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import ErrorDisplay from '@/components/ErrorDisplay.vue'
import { useErrorHandler } from '@/composables/useErrorHandler'
import { useToast } from '@/composables/useToast'
import AdminTaskTemplatesView from '@/views/AdminTaskTemplatesView.vue'
import type { User } from '@/types'

interface ListUsersResponse {
  users: User[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

const { t } = useI18n()
const toast = useToast()
const users = ref<User[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const totalPages = ref(0)
const searchQuery = ref('')
const loading = ref(true)
const activeTab = ref<'users' | 'storage' | 'task-templates'>('users')
const settings = ref<Record<string, string>>({})
const saved = ref(false)
const showDefaultPwd = ref(false)
const { error, setError, clearError } = useErrorHandler()

// ── Reset password dialog ──────────────────────────────────────
const resetTarget = ref<User | null>(null)
const resetSubmitting = ref(false)
const { error: resetError, setError: setResetError, clearError: clearResetError } = useErrorHandler()

onMounted(async () => {
  loading.value = true
  await fetchUsers()
  await loadSettings()
  loading.value = false
})

async function fetchUsers() {
  try {
    const params = new URLSearchParams()
    if (searchQuery.value) params.set('q', searchQuery.value)
    params.set('page', String(page.value))
    params.set('page_size', String(pageSize))

    const data = await api.get<ListUsersResponse>(`/admin/users?${params}`)
    users.value = data.users
    total.value = data.total
    totalPages.value = data.total_pages
  } catch { /* */ }
}

// Debounced search
let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(searchQuery, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    fetchUsers()
  }, 300)
})

function goToPage(p: number) {
  if (p < 1 || p > totalPages.value) return
  page.value = p
  fetchUsers()
}

// ── Reset password ─────────────────────────────────────────────
function openResetDialog(u: User) {
  resetTarget.value = u
  clearResetError()
}

function closeResetDialog() {
  resetTarget.value = null
}

async function submitResetPassword() {
  if (!resetTarget.value || resetSubmitting.value) return
  clearResetError()
  resetSubmitting.value = true
  try {
    const res = await api.post<{ password: string }>('/admin/users/reset-password', {
      user_id: resetTarget.value.id,
    })
    toast.success(t('admin.resetSuccess'))
    // Show the password that was set
    toast.info(`${t('admin.resetPasswordIs')}: ${res.password}`)
    closeResetDialog()
  } catch (e: any) {
    setResetError(e)
  } finally {
    resetSubmitting.value = false
  }
}

// ── Default password management ────────────────────────────────
function regeneratePassword() {
  const chars = 'abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789'
  let pwd = ''
  for (let i = 0; i < 12; i++) {
    pwd += chars[Math.floor(Math.random() * chars.length)]
  }
  settings.value['default_password'] = pwd
}

async function saveDefaultPassword() {
  clearError()
  saved.value = false
  try {
    await api.put('/admin/settings', { default_password: settings.value['default_password'] })
    saved.value = true
    setTimeout(() => { saved.value = false }, 2000)
  } catch (e: any) { setError(e) }
}

async function loadSettings() {
  try {
    const list = await api.get<{ Key: string; Value: string }[]>('/admin/settings')
    const map: Record<string, string> = {}
    for (const s of list) { map[s.Key] = s.Value }
    settings.value = map
  } catch { /* */ }
}

async function saveSettings() {
  clearError()
  saved.value = false
  try {
    await api.put('/admin/settings', settings.value)
    saved.value = true
    setTimeout(() => { saved.value = false }, 2000)
  } catch (e: any) { setError(e) }
}

const STORAGE_OPTIONS = [
  { value: 'local', label: '本地存储 (Local)' },
  { value: 's3', label: 'AWS S3（预留）', disabled: true },
  { value: 'oss', label: '阿里云 OSS（预留）', disabled: true },
  { value: 'minio', label: 'MinIO（预留）', disabled: true },
]

const pages = computed(() => {
  const result: number[] = []
  const start = Math.max(1, page.value - 2)
  const end = Math.min(totalPages.value, page.value + 2)
  for (let i = start; i <= end; i++) result.push(i)
  return result
})

const showingText = computed(() => {
  const from = (page.value - 1) * pageSize + 1
  const to = Math.min(page.value * pageSize, total.value)
  return `${t('admin.showingFrom')} ${from}-${to} ${t('admin.showingOf')} ${total.value}`
})
</script>

<template>
  <div class="w-full max-w-5xl mx-auto p-4">
    <h2 class="text-xl md:text-2xl font-bold mb-4 dark:text-gray-200">{{ t('admin.heading') }}</h2>

    <!-- Tabs -->
    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'users' }"
        @click="activeTab = 'users'"
      >{{ t('admin.users') }}</button>
      <button class="tab" :class="{ active: activeTab === 'storage' }"
        @click="activeTab = 'storage'"
      >{{ t('admin.storage') }}</button>
      <button class="tab" :class="{ active: activeTab === 'task-templates' }"
        @click="activeTab = 'task-templates'"
      >{{ t('taskTemplate.heading') }}</button>
    </div>

    <!-- Users Tab -->
    <div v-if="activeTab === 'users'" class="card overflow-x-auto">
      <!-- Search -->
      <div class="mb-3">
        <input
          v-model="searchQuery"
          type="text"
          class="input max-w-xs"
          :placeholder="t('admin.searchPlaceholder')"
        />
      </div>

      <table class="w-full text-sm min-w-[600px]">
          <thead>
            <tr class="border-b dark:border-gray-700 text-left text-gray-500 dark:text-gray-400">
              <th class="py-2 px-3">{{ t('admin.usersDisplayName') }}</th>
              <th class="py-2 px-3">{{ t('admin.usersEmail') }}</th>
              <th class="py-2 px-3">{{ t('admin.usersRoles') }}</th>
              <th class="py-2 px-3 hidden sm:table-cell">{{ t('admin.usersCreated') }}</th>
              <th class="py-2 px-3">{{ t('admin.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in users" :key="u.id" class="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
              <td class="py-2 px-3 font-medium dark:text-gray-200">{{ u.display_name }}</td>
              <td class="py-2 px-3 dark:text-gray-300">{{ u.email }}</td>
              <td class="py-2 px-3">
                <span v-if="u.roles.includes('admin')" class="text-primary font-medium">{{ t('user.admin') }}</span>
                <span v-else class="text-gray-400">{{ t('user.member') }}</span>
              </td>
              <td class="py-2 px-3 text-gray-400 hidden sm:table-cell">{{ u.created_at?.split('T')[0] }}</td>
              <td class="py-2 px-3">
                <button
                  class="text-xs text-blue-500 hover:underline"
                  @click="openResetDialog(u)"
                >{{ t('admin.resetPassword') }}</button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Empty state -->
        <p v-if="!loading && users.length === 0" class="text-center text-gray-400 py-6">
          {{ searchQuery ? t('admin.noResults') : t('admin.noUsers') }}
        </p>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="flex items-center justify-between pt-3 text-sm">
          <span class="text-gray-400">{{ showingText }}</span>
          <div class="flex gap-1">
            <button class="px-2 py-1 border rounded dark:border-gray-600 disabled:opacity-30" :disabled="page <= 1" @click="goToPage(page - 1)">‹</button>
            <button
              v-for="p in pages" :key="p"
              class="px-2 py-1 border rounded dark:border-gray-600"
              :class="p === page ? 'bg-primary text-white border-primary' : ''"
              @click="goToPage(p)"
            >{{ p }}</button>
            <button class="px-2 py-1 border rounded dark:border-gray-600 disabled:opacity-30" :disabled="page >= totalPages" @click="goToPage(page + 1)">›</button>
          </div>
        </div>
      </div>

    <!-- Storage Tab -->
    <div v-if="activeTab === 'storage'" class="card space-y-4">
      <ErrorDisplay :error="error" @close="clearError" />
      <LoadingSpinner :text="t('app.loading')" v-if="loading" />
      <template v-else>
        <section>
          <h3 class="font-medium mb-3 dark:text-gray-200">{{ t('admin.storageHeading') }}</h3>
          <p class="text-xs text-gray-400 mb-4">{{ t('admin.storageDesc') }}</p>
          <label class="block text-sm text-gray-500 dark:text-gray-400 mb-1">{{ t('admin.storageType') }}</label>
          <select v-model="settings['storage.type']" class="input mb-4 max-w-xs">
            <option v-for="opt in STORAGE_OPTIONS" :key="opt.value" :value="opt.value" :disabled="opt.disabled">{{ opt.label }}</option>
          </select>
          <button class="btn-primary text-sm" @click="saveSettings">
            {{ saved ? t('admin.storageSaved') : t('admin.storageSave') }}
          </button>
        </section>

        <section class="border-t dark:border-gray-700 pt-4">
          <h3 class="font-medium mb-2 dark:text-gray-200">{{ t('admin.currentStatus') }}</h3>
          <div class="text-sm text-gray-500 dark:text-gray-400 space-y-1">
            <p>{{ t('admin.storageTypeLabel') }}<code class="bg-gray-100 dark:bg-gray-700 px-1.5 py-0.5 rounded text-xs">{{ settings['storage.type'] || 'local' }}</code></p>
            <p v-if="settings['storage.type'] === 'local' || !settings['storage.type']" class="text-xs text-green-600 mt-2">{{ t('admin.storageLocalActive') }}</p>
          </div>
        </section>

        <!-- Default Password -->
        <section class="border-t dark:border-gray-700 pt-4">
          <h3 class="font-medium mb-3 dark:text-gray-200">{{ t('admin.defaultPasswordHeading') }}</h3>
          <p class="text-xs text-gray-400 mb-4">{{ t('admin.defaultPasswordDesc') }}</p>
          <div class="flex items-center gap-2 mb-3 max-w-md">
            <input
              :type="showDefaultPwd ? 'text' : 'password'"
              :value="settings['default_password'] || ''"
              readonly
              class="input flex-1 font-mono text-sm"
            />
            <button class="btn-secondary text-xs px-3 py-2" @click="showDefaultPwd = !showDefaultPwd">
              {{ showDefaultPwd ? t('admin.hide') : t('admin.show') }}
            </button>
          </div>
          <div class="flex gap-2">
            <button class="btn-secondary text-sm" @click="regeneratePassword">{{ t('admin.regenerate') }}</button>
            <button class="btn-primary text-sm" @click="saveDefaultPassword">
              {{ saved ? t('admin.storageSaved') : t('admin.storageSave') }}
            </button>
          </div>
        </section>
      </template>
    </div>

    <!-- Task Templates Tab -->
    <div v-if="activeTab === 'task-templates'" class="card">
      <AdminTaskTemplatesView />
    </div>

    <!-- ── Reset Password Modal ──────────────────────────────── -->
    <Teleport to="body">
      <div v-if="resetTarget" class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" @click.self="closeResetDialog">
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-sm p-6">
          <h3 class="text-lg font-semibold mb-2 dark:text-gray-100">{{ t('admin.resetPasswordTitle') }}</h3>
          <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
            {{ t('admin.resetPasswordConfirm', { user: resetTarget.display_name }) }}
          </p>
          <ErrorDisplay :error="resetError" @close="clearResetError" />
          <div class="flex justify-end gap-2 mt-3">
            <button class="btn-secondary text-sm" @click="closeResetDialog">{{ t('confirm.cancel') }}</button>
            <button class="btn-primary text-sm" :disabled="resetSubmitting" @click="submitResetPassword">
              {{ resetSubmitting ? '...' : t('admin.resetConfirm') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
