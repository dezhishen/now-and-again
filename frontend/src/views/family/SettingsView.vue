<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>

<template>
  <div>
    <h2 class="text-xl md:text-2xl font-bold mb-4 dark:text-gray-200">{{ t('settingsPage.heading') }}</h2>
    <p class="text-gray-400 dark:text-gray-500">TODO</p>
  </div>
</template>
