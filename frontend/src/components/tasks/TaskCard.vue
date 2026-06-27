<script setup lang="ts">
import type { Task } from '@/types'
import { getTaskCard } from '@/composables/useTaskKinds'

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
  <component
    :is="getTaskCard(task.kind)"
    :task="task"
    :loc-name="locName"
    :loc-color="locColor"
    :group-name="groupName"
    :summary="summary"
    @edit="$emit('edit', $event)"
    @logs="$emit('logs', $event)"
    @trigger="$emit('trigger', $event)"
    @toggle="$emit('toggle', $event)"
    @delete="$emit('delete', $event)"
  />
</template>
