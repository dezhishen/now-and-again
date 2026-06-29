<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from '@/i18n'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const route = useRoute()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const submitting = ref(false)
const error = ref('')

const redirect = (() => {
  const q = route.query.redirect as string
  if (q && q !== '/login' && q !== '/register') return q
  return '/'
})()

async function handleLogin() {
  if (submitting.value) return
  error.value = ''
  submitting.value = true
  try {
    await auth.login(username.value, password.value)
    // Use hard navigation — Vue Router guards may conflict with auth state changes.
    window.location.href = redirect
  } catch (e: any) {
    error.value = e.message || t('login.error')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-screen px-4">
    <div class="card w-full max-w-md">
      <h1 class="text-2xl font-bold text-center mb-6 dark:text-gray-200">{{ t('login.heading') }}</h1>
      <form @submit.prevent="handleLogin" class="flex flex-col gap-3">
        <input
          v-model="username"
          class="input"
          :placeholder="t('login.usernamePlaceholder')"
          autocomplete="username"
          required
        />
        <input
          v-model="password"
          type="password"
          class="input"
          :placeholder="t('login.passwordPlaceholder')"
          autocomplete="current-password"
          required
        />
        <p v-if="error" class="text-danger text-sm">{{ error }}</p>
        <button type="submit" class="btn-primary w-full mt-2" :disabled="submitting">
          {{ submitting ? '...' : t('login.submit') }}
        </button>
      </form>
      <p class="text-center text-sm text-gray-400 dark:text-gray-500 mt-4">
        {{ t('login.noAccount') }}<router-link to="/register" class="text-primary hover:underline ml-1">{{ t('login.toRegister') }}</router-link>
      </p>
    </div>
  </div>
</template>
