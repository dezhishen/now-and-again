<script setup lang="ts">
import type { FamilyGroup } from '@/types'
import SubTaskEditor from '@/components/tasks/SubTaskEditor.vue'

const steps = defineModel<any[]>({ required: true })
const props = defineProps<{
  groups: FamilyGroup[]
  locations: { id: string; name: string }[]
}>()

function addStep() {
  steps.value.push({
    task: {
      task: {
        name: '',
        kind: 'simple',
        schedule_type: 'once',
        schedule_data: { time: '09:00' },
        group_id: '',
        location_id: '',
      },
    },
  })
}

function removeStep(index: number) {
  steps.value.splice(index, 1)
}

function moveUp(index: number) {
  if (index > 0) {
    const tmp = steps.value[index - 1]
    steps.value[index - 1] = steps.value[index]
    steps.value[index] = tmp
  }
}

function moveDown(index: number) {
  if (index < steps.value.length - 1) {
    const tmp = steps.value[index + 1]
    steps.value[index + 1] = steps.value[index]
    steps.value[index] = tmp
  }
}
</script>

<template>
  <div class="space-y-3 border-l-2 border-orange-400 pl-3">
    <div class="flex items-center justify-between">
      <p class="text-xs text-orange-600 dark:text-orange-400 font-medium">🔗 任务步骤</p>
      <button class="text-xs text-primary hover:underline" data-testid="chain-add-step" @click="addStep">+ 添加步骤</button>
    </div>
    <div class="max-h-80 overflow-y-auto space-y-2">
      <div v-for="(step, i) in steps" :key="i" class="space-y-1 pb-2 border-b border-gray-100 dark:border-gray-700 last:border-0">
        <div class="flex items-center gap-2">
          <span class="text-xs text-gray-400 font-mono flex-shrink-0">{{ i + 1 }}.</span>
          <SubTaskEditor
            v-if="step.task"
            v-model="step.task"
            :groups="props.groups"
            :locations="props.locations"
            class="flex-1"
          />
          <div class="flex flex-col gap-0.5 flex-shrink-0">
            <button class="text-xs text-gray-400 hover:text-gray-600" @click="moveUp(i)" :disabled="i === 0">↑</button>
            <button class="text-xs text-gray-400 hover:text-gray-600" @click="moveDown(i)" :disabled="i === steps.length - 1">↓</button>
          </div>
          <button class="text-xs text-danger hover:underline flex-shrink-0" @click="removeStep(i)">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>
