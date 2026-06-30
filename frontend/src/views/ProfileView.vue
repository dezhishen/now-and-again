<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from '@/i18n'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import ErrorDisplay from '@/components/ErrorDisplay.vue'
import { useErrorHandler } from '@/composables/useErrorHandler'

const { t } = useI18n()
const auth = useAuthStore()
const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const { error, setError, clearError } = useErrorHandler()

const displayName = ref('')
const email = ref('')
const phone = ref('')

onMounted(async () => {
  loading.value = true
  try {
    const user = await api.get<{ display_name: string; email: string; phone: string }>('/users/me')
    displayName.value = user.display_name || ''
    email.value = user.email || ''
    phone.value = user.phone || ''
  } catch { /* */ }
  loading.value = false
})

async function save() {
  saving.value = true
  clearError()
  saved.value = false
  try {
    const body: Record<string, string> = {}
    if (displayName.value !== (auth.user?.display_name || '')) body.display_name = displayName.value
    if (email.value !== (auth.user?.email || '')) body.email = email.value
    if (phone.value !== (auth.user?.phone || '')) body.phone = phone.value
    await api.put('/users/me', body)
    // Refresh local user state
    if (auth.user) {
      if (body.display_name) auth.user.display_name = body.display_name
      if (body.email) auth.user.email = body.email
      if (body.phone) auth.user.phone = body.phone
    }
    saved.value = true
    setTimeout(() => { saved.value = false }, 2000)
  } catch (e: any) { setError(e) }
  finally { saving.value = false }
}
</script>

<template>
  <div class="max-w-2xl mx-auto p-4">
    <h2 class="text-xl md:text-2xl font-bold mb-6 dark:text-gray-200">{{ t('profile.heading') }}</h2>

    <ErrorDisplay :error="error" @close="clearError" />
    <LoadingSpinner :text="t('app.loading')" v-if="loading" />

    <template v-else>
      <div class="card mb-6">
        <div class="flex items-center gap-4 mb-6">
          <div class="w-16 h-16 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-2xl flex-shrink-0">
            {{ auth.user?.display_name?.[0]?.toUpperCase() || '?' }}
          </div>
          <div>
            <p class="font-medium text-lg dark:text-gray-200">{{ auth.user?.display_name }}</p>
            <p class="text-sm text-gray-400">{{ auth.user?.email }}</p>
            <p v-if="auth.isAdmin" class="text-xs text-primary mt-0.5">{{ t('profile.adminRole') }}</p>
          </div>
        </div>

        <div class="space-y-4">
          <div>
            <label class="text-xs text-gray-400 block mb-1">{{ t('profile.displayName') }}</label>
            <input v-model="displayName" class="input" :placeholder="t('profile.displayNamePlaceholder')" />
          </div>
          <div>
            <label class="text-xs text-gray-400 block mb-1">{{ t('profile.email') }}</label>
            <input v-model="email" type="email" class="input" placeholder="your@email.com" />
          </div>
          <div>
            <label class="text-xs text-gray-400 block mb-1">{{ t('profile.phone') }}</label>
            <input v-model="phone" class="input" :placeholder="t('profile.phoneOptional')" />
          </div>
        </div>
      </div>

      <button class="btn-primary" :disabled="saving" @click="save">
        {{ saving ? t('profile.saving') : saved ? t('profile.saved') : t('profile.save') }}
      </button>
    </template>
  </div>
</template>
