<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>

<template>
  <div class="flex flex-col items-center justify-center min-h-[60vh] gap-3">
    <h1 class="text-6xl font-bold text-gray-300 dark:text-gray-600">{{ t('notFound.code') }}</h1>
    <p class="text-lg dark:text-gray-400">{{ t('notFound.message') }}</p>
    <router-link to="/" class="text-primary hover:underline">{{ t('notFound.backHome') }}</router-link>
  </div>
</template>
