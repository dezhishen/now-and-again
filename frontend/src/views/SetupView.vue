<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const form = ref({ username: '', email: '', password: '', display_name: '' })

async function handleSetup() {
  const ok = await auth.setup(form.value)
  if (ok) router.push('/login')
}
</script>

<template>
  <div class="flex flex-col items-center justify-center min-h-screen bg-gray-50 px-4">
    <div class="card w-full max-w-md">
      <h1 class="text-2xl font-bold text-center mb-1">{{ t('setup.title') }}</h1>
      <p class="text-center text-muted text-sm mb-6">{{ t('setup.heading') }}</p>
      <form @submit.prevent="handleSetup" class="flex flex-col gap-3">
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('setup.displayName') }}</label>
          <input v-model="form.display_name" class="input" required />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('setup.username') }}</label>
          <input v-model="form.username" class="input" :placeholder="t('setup.usernamePlaceholder')" required minlength="3" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('setup.email') }}</label>
          <input v-model="form.email" type="email" class="input" :placeholder="t('setup.emailPlaceholder')" required />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('setup.password') }}</label>
          <input v-model="form.password" type="password" class="input" :placeholder="t('setup.passwordPlaceholder')" required minlength="8" />
        </div>
        <p v-if="auth.error" class="text-danger text-sm">{{ auth.error }}</p>
        <button type="submit" class="btn-primary w-full mt-2">{{ t('setup.submit') }}</button>
      </form>
    </div>
  </div>
</template>
