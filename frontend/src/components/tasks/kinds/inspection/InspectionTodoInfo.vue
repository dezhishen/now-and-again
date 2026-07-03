<script setup lang="ts">
import type { Todo } from '@/types'

defineProps<{ todo: Todo }>()
</script>

<template>
  <p class="text-xs text-purple-400 flex items-center gap-1 h-5">
    <template v-if="todo.task?.display_summary">
      <span class="flex-shrink-0">📋</span>
      <span class="truncate" :title="todo.task.display_summary">{{ todo.task.display_summary }}</span>
    </template>
    <span v-else class="invisible">.</span>
  </p>
</template>
