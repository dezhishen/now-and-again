<script setup lang="ts">
import { useI18n } from '@/i18n'
const { t } = useI18n()
defineProps<{ text?: string }>()
</script>

<template>
  <div class="flex items-center justify-center py-12">
    <div class="flex flex-col items-center gap-3">
      <div class="w-8 h-8 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
      <span class="text-sm text-gray-400">{{ text || t('app.loading') }}</span>
    </div>
  </div>
</template>
