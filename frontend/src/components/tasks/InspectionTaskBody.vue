<script setup lang="ts">
import type { TaskTemplate, CheckItem } from '@/types'

defineProps<{
  task: TaskTemplate
  locName: (id: string) => string
  locColor: (id: string) => string
  groupName: (id: string) => string
  summary: (t: TaskTemplate) => string
}>()
</script>

<template>
  <div v-if="(task.check_items as CheckItem[])?.length" class="text-xs text-gray-400 mt-1 space-y-0.5">
    <div v-for="item in (task.check_items as CheckItem[])" :key="item.name">
      <span class="font-medium text-gray-500">🔍 {{ item.name }}:</span>
      <span class="flex flex-wrap gap-0.5 ml-1">
        <span v-for="b in item.branches" :key="b.name" class="px-1 rounded" :class="b.create_todo ? 'text-warning' : 'text-green-600'">
          {{ b.create_todo ? '⚠' : '✓' }}{{ b.name }}
        </span>
      </span>
    </div>
  </div>
</template>
