<script setup lang="ts">
import { useToast } from '@/composables/useToast'

const { toasts, remove } = useToast()

const ICONS: Record<string, string> = {
  success: '✅',
  error: '❌',
  warning: '⚠️',
  info: 'ℹ️',
}
const BG: Record<string, string> = {
  success: 'bg-green-50 dark:bg-green-900/30 border-green-200 dark:border-green-800 text-green-800 dark:text-green-200',
  error: 'bg-red-50 dark:bg-red-900/30 border-red-200 dark:border-red-800 text-red-800 dark:text-red-200',
  warning: 'bg-yellow-50 dark:bg-yellow-900/30 border-yellow-200 dark:border-yellow-800 text-yellow-800 dark:text-yellow-200',
  info: 'bg-blue-50 dark:bg-blue-900/30 border-blue-200 dark:border-blue-800 text-blue-800 dark:text-blue-200',
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed top-4 right-4 z-[100] flex flex-col gap-2 max-w-sm">
      <div v-for="t in toasts" :key="t.id"
        class="flex items-center gap-2 px-4 py-3 rounded-lg border shadow-lg text-sm transition-all cursor-pointer"
        :class="BG[t.level]"
        @click="remove(t.id)"
      >
        <span>{{ ICONS[t.level] }}</span>
        <span class="flex-1">{{ t.message }}</span>
      </div>
    </div>
  </Teleport>
</template>
