<script setup lang="ts">
import type { Todo } from '@/types'

defineProps<{ todo: Todo }>()
</script>

<template>
  <p v-if="todo.task?.display_summary" class="text-xs text-purple-400 flex items-center gap-1">
    <span>📋</span>
    <span>{{ todo.task.display_summary }}</span>
  </p>
</template>
