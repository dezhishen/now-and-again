<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const username = ref('')
const password = ref('')
const submitting = ref(false)

async function handleLogin() {
  if (submitting.value) return
  submitting.value = true
  try {
    const ok = await auth.login(username.value, password.value)
    if (ok) router.push('/')
  } finally { submitting.value = false }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-screen px-4">
    <div class="card w-full max-w-md">
      <h1 class="text-2xl font-bold text-center mb-6 dark:text-gray-200">{{ t('login.heading') }}</h1>
      <form @submit.prevent="handleLogin" class="flex flex-col gap-3">
        <input v-model="username" class="input" :placeholder="t('login.usernamePlaceholder')" required />
        <input v-model="password" type="password" class="input" :placeholder="t('login.passwordPlaceholder')" required />
        <p v-if="auth.error" class="text-danger text-sm">{{ auth.error }}</p>
        <button type="submit" class="btn-primary w-full mt-2" :disabled="submitting">{{ submitting ? '...' : t('login.submit') }}</button>
      </form>
      <p class="text-center text-sm text-gray-400 dark:text-gray-500 mt-4">
        {{ t('login.noAccount') }}<router-link to="/register">{{ t('login.toRegister') }}</router-link>
      </p>
    </div>
  </div>
</template>
