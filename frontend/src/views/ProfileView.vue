<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '@/i18n'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import ErrorDisplay from '@/components/ErrorDisplay.vue'
import { useErrorHandler } from '@/composables/useErrorHandler'

const { t } = useI18n()
const auth = useAuthStore()
const toast = useToast()
const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const { error, setError, clearError } = useErrorHandler()

const displayName = ref('')
const email = ref('')
const phone = ref('')

// ── Change password dialog ──────────────────────────────────────
const showPwdDialog = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const pwdSubmitting = ref(false)
const { error: pwdError, setError: setPwdError, clearError: clearPwdError } = useErrorHandler()

const pwdMismatch = computed(() =>
  confirmPassword.value !== '' && newPassword.value !== confirmPassword.value
)

const canSubmitPwd = computed(() =>
  oldPassword.value.length > 0 &&
  newPassword.value.length >= 8 &&
  confirmPassword.value.length > 0 &&
  !pwdMismatch.value
)

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

function openPwdDialog() {
  oldPassword.value = ''
  newPassword.value = ''
  confirmPassword.value = ''
  clearPwdError()
  showPwdDialog.value = true
}

function closePwdDialog() {
  showPwdDialog.value = false
}

async function submitPassword() {
  if (!canSubmitPwd.value || pwdSubmitting.value) return
  clearPwdError()
  pwdSubmitting.value = true
  try {
    await api.put('/users/me/password', {
      old_password: oldPassword.value,
      new_password: newPassword.value,
    })
    toast.success(t('profile.passwordChanged'))
    closePwdDialog()
  } catch (e: any) {
    setPwdError(e)
  } finally {
    pwdSubmitting.value = false
  }
}
</script>

<template>
  <div class="max-w-3xl mx-auto p-4">
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

      <div class="flex gap-3">
        <button class="btn-primary" :disabled="saving" @click="save">
          {{ saving ? t('profile.saving') : saved ? t('profile.saved') : t('profile.save') }}
        </button>
        <button class="btn-secondary" @click="openPwdDialog">
          {{ t('profile.changePassword') }}
        </button>
      </div>
    </template>

    <!-- ── Change Password Modal ──────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showPwdDialog" class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" @click.self="closePwdDialog">
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-sm p-6">
          <h3 class="text-lg font-semibold mb-4 dark:text-gray-100">{{ t('profile.changePassword') }}</h3>

          <div class="space-y-3">
            <div>
              <label class="text-xs text-gray-400 block mb-1">{{ t('profile.oldPassword') }}</label>
              <input v-model="oldPassword" type="password" class="input" autocomplete="current-password" />
            </div>
            <div>
              <label class="text-xs text-gray-400 block mb-1">{{ t('profile.newPassword') }}</label>
              <input v-model="newPassword" type="password" class="input" minlength="8" autocomplete="new-password" />
            </div>
            <div>
              <label class="text-xs text-gray-400 block mb-1">{{ t('profile.confirmPassword') }}</label>
              <input v-model="confirmPassword" type="password" class="input" autocomplete="new-password" />
              <p v-if="pwdMismatch" class="text-xs text-red-500 mt-1">{{ t('profile.passwordMismatch') }}</p>
            </div>
          </div>

          <ErrorDisplay :error="pwdError" @close="clearPwdError" />
          <div class="flex justify-end gap-2 mt-4">
            <button class="btn-secondary text-sm" @click="closePwdDialog">{{ t('confirm.cancel') }}</button>
            <button class="btn-primary text-sm" :disabled="!canSubmitPwd || pwdSubmitting" @click="submitPassword">
              {{ pwdSubmitting ? t('app.submitting') : t('profile.confirmChange') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
