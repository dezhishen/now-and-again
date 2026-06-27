<script setup lang="ts">
import type { TaskTemplate } from '@/types'

defineProps<{
  task: TaskTemplate
  locName: (id: string) => string
  locColor: (id: string) => string
  groupName: (id: string) => string
  summary: (t: TaskTemplate) => string
}>()

defineEmits<{
  edit: [task: TaskTemplate]
  logs: [id: string]
  trigger: [id: string]
  toggle: [task: TaskTemplate]
  delete: [id: string]
}>()
</script>

<template>
  <div class="card hover:shadow-md transition-shadow">
    <div class="flex items-start justify-between mb-2">
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <span class="font-medium dark:text-gray-200 truncate">{{ task.name }}</span>
          <span class="text-[10px] px-1.5 py-0.5 rounded-full bg-purple-100 dark:bg-purple-900/40 text-purple-600 dark:text-purple-400 flex-shrink-0">巡检</span>
          <span class="flex-shrink-0 w-1.5 h-1.5 rounded-full" :class="task.enabled ? 'bg-green-500' : 'bg-gray-300'" />
        </div>
        <div class="flex items-center justify-between gap-2 mt-1 h-5">
          <span class="text-xs text-gray-400 truncate">{{ summary(task) }}</span>
          <span v-if="task.location_id" class="text-xs px-1.5 py-0.5 rounded flex-shrink-0" :style="{ background: locColor(task.location_id) + '20', color: locColor(task.location_id) }">
            📍 {{ locName(task.location_id) }}
          </span>
        </div>
        <div v-if="task.group_id" class="flex items-center gap-1 mt-1">
          <span class="text-xs text-gray-400">👥 {{ groupName(task.group_id) }}</span>
        </div>
        <p v-if="task.display_summary" class="text-xs text-purple-400 mt-1">🔍 {{ task.display_summary }}</p>
      </div>
    </div>
    <div class="flex gap-1 border-t dark:border-gray-700 pt-2 mt-2">
      <button class="text-xs px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400 flex-1" @click="$emit('edit', task)">编辑</button>
      <button class="text-xs px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400 flex-1" @click="$emit('logs', task.id)">日志</button>
      <button class="text-xs px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400 flex-1" @click="$emit('trigger', task.id)">生成</button>
      <button class="text-xs px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400 flex-1" @click="$emit('toggle', task)">{{ task.enabled ? '禁用' : '启用' }}</button>
      <button class="text-xs px-2 py-0.5 rounded text-danger hover:bg-red-50 dark:hover:bg-red-900/30 flex-1" @click="$emit('delete', task.id)">删除</button>
    </div>
  </div>
</template>
