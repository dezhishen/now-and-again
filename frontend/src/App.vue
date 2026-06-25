<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useDark } from '@/composables/useDark'

const { isDark, toggle } = useDark()
</script>

<template>
  <div id="na-app" class="min-h-screen">
    <!-- Theme toggle -->
    <button
      class="fixed top-3 right-3 z-50 w-9 h-9 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center text-sm shadow hover:scale-110 transition-transform"
      @click="toggle"
      :title="isDark ? '切换亮色' : '切换暗色'"
    >
      {{ isDark ? '☀️' : '🌙' }}
    </button>
    <RouterView />
  </div>
</template>
