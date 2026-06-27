<script setup lang="ts">
import type { TaskTemplate } from '@/types'
import { getTaskCard } from '@/composables/useTaskKinds'

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
