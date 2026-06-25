<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const form = ref({ username: '', email: '', password: '', display_name: '' })

async function handleRegister() { if (await auth.register(form.value)) router.push('/login') }
</script>

<template>
  <div class="flex items-center justify-center min-h-screen px-4">
    <div class="card w-full max-w-md">
      <h1 class="text-2xl font-bold text-center mb-6 dark:text-gray-200">{{ t('register.heading') }}</h1>
      <form @submit.prevent="handleRegister" class="flex flex-col gap-3">
        <input v-model="form.display_name" class="input" :placeholder="t('register.displayNamePlaceholder')" required />
        <input v-model="form.username" class="input" :placeholder="t('register.usernamePlaceholder')" required minlength="3" />
        <input v-model="form.email" type="email" class="input" :placeholder="t('register.emailPlaceholder')" required />
        <input v-model="form.password" type="password" class="input" :placeholder="t('register.passwordPlaceholder')" required minlength="8" />
        <p v-if="auth.error" class="text-danger text-sm">{{ auth.error }}</p>
        <button type="submit" class="btn-primary w-full mt-2">{{ t('register.submit') }}</button>
      </form>
      <p class="text-center text-sm text-gray-400 dark:text-gray-500 mt-4">
        {{ t('register.hasAccount') }}<router-link to="/login">{{ t('register.toLogin') }}</router-link>
      </p>
    </div>
  </div>
</template>
