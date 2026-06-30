<script setup lang="ts">
defineProps<{ text?: string }>()
</script>

<template>
  <div class="flex items-center justify-center py-12">
    <div class="flex flex-col items-center gap-3">
      <div class="w-8 h-8 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
      <span v-if="text" class="text-sm text-gray-400">{{ text }}</span>
    </div>
  </div>
</template>
