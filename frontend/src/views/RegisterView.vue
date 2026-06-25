<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const form = ref({ username: '', email: '', password: '', display_name: '' })

async function handleRegister() {
  const ok = await auth.register(form.value)
  if (ok) router.push('/login')
}
</script>

<template>
  <div class="flex flex-col items-center justify-center min-h-screen bg-gray-50 px-4">
    <div class="card w-full max-w-md">
      <h1 class="text-2xl font-bold text-center mb-6">{{ t('register.heading') }}</h1>
      <form @submit.prevent="handleRegister" class="flex flex-col gap-3">
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('register.displayName') }}</label>
          <input v-model="form.display_name" class="input" :placeholder="t('register.displayNamePlaceholder')" required />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('register.username') }}</label>
          <input v-model="form.username" class="input" :placeholder="t('register.usernamePlaceholder')" required minlength="3" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('register.email') }}</label>
          <input v-model="form.email" type="email" class="input" :placeholder="t('register.emailPlaceholder')" required />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('register.password') }}</label>
          <input v-model="form.password" type="password" class="input" :placeholder="t('register.passwordPlaceholder')" required minlength="8" />
        </div>
        <p v-if="auth.error" class="text-danger text-sm">{{ auth.error }}</p>
        <button type="submit" class="btn-primary w-full mt-2">{{ t('register.submit') }}</button>
      </form>
      <p class="text-center text-sm text-muted mt-4">
        {{ t('register.hasAccount') }}<router-link to="/login">{{ t('register.toLogin') }}</router-link>
      </p>
    </div>
  </div>
</template>
