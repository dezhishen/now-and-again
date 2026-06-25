<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>

<template>
  <div>
    <h2 class="text-xl md:text-2xl font-bold mb-4 dark:text-gray-200">{{ t('dashboard.heading') }}</h2>
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
      <div class="card"><p class="text-gray-400 text-sm">{{ t('dashboard.todoTasks') }}</p><p class="text-2xl font-bold dark:text-gray-200">0</p></div>
      <div class="card"><p class="text-gray-400 text-sm">{{ t('dashboard.inProgress') }}</p><p class="text-2xl font-bold dark:text-gray-200">0</p></div>
      <div class="card"><p class="text-gray-400 text-sm">{{ t('dashboard.completed') }}</p><p class="text-2xl font-bold dark:text-gray-200">0</p></div>
    </div>
  </div>
</template>
