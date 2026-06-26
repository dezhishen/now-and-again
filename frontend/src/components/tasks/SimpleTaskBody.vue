<script setup lang="ts">
import type { TaskTemplate } from '@/types'

defineProps<{
  task: TaskTemplate
  locName: (id: string) => string
  locColor: (id: string) => string
  groupName: (id: string) => string
  summary: (t: TaskTemplate) => string
}>()
</script>

<template>
  <!-- Simple task: no extra body content -->
</template>
