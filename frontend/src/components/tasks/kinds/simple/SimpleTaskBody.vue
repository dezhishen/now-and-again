<script setup lang="ts">
import type { Task } from '@/types'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps<{
  task: Task
  locName: (id: string) => string
  locColor: (id: string) => string
  groupName: (id: string) => string
  summary: (t: Task) => string
}>()

defineEmits<{
  edit: [task: Task]
  logs: [id: string]
  trigger: [id: string]
  toggle: [task: Task]
  delete: [id: string]
}>()
</script>

<template>
  <div class="card hover:shadow-md transition-shadow relative overflow-hidden">
    <!-- Kind ribbon -->
    <div class="absolute -top-0.5 -right-0.5 w-14 h-14 overflow-hidden z-10">
      <div class="absolute top-2.5 -right-[18px] w-16 bg-blue-400 text-white text-[10px] font-medium text-center leading-4 rotate-45 shadow-sm">{{ t('taskCard.simpleKind') }}</div>
    </div>
    <div class="flex items-start justify-between mb-2">
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <span class="font-medium dark:text-gray-200 truncate">{{ task.name }}</span>
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
        <p v-else class="text-xs mt-1 invisible">.</p>
      </div>
    </div>
    <div class="flex gap-1 border-t dark:border-gray-700 pt-2 mt-2">
      <button class="btn-ghost text-xs flex-1" @click="$emit('edit', task)">{{ t('taskCard.edit') }}</button>
      <button class="btn-ghost text-xs flex-1" @click="$emit('logs', task.id)">{{ t('taskCard.logs') }}</button>
      <button class="btn-ghost text-xs flex-1" @click="$emit('trigger', task.id)">{{ t('taskCard.trigger') }}</button>
      <button class="btn-ghost text-xs flex-1" @click="$emit('toggle', task)">{{ task.enabled ? t('taskCard.disable') : t('taskCard.enable') }}</button>
      <button class="text-xs px-2 py-0.5 rounded text-danger hover:bg-red-50 dark:hover:bg-red-900/30 flex-1" @click="$emit('delete', task.id)">{{ t('taskCard.delete') }}</button>
    </div>
  </div>
</template>
