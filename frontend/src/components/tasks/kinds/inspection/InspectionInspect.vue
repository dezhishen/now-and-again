<script setup lang="ts">
import type { Task } from '@/types'

const task = defineModel<Task | null>('task', { required: true })
const selections = defineModel<Record<string, string>>('selections', { default: () => ({}) })

function selectBranch(itemName: string, branchName: string) {
  selections.value[itemName] = branchName
}
</script>

<template>
  <div class="flex-1 overflow-auto p-4 space-y-4">
    <div v-for="item in (task?.extra?.check_items || [])" :key="item.name" class="space-y-1">
      <p class="text-sm font-medium text-gray-600 dark:text-gray-300">{{ item.name }}</p>
      <div class="flex flex-wrap gap-1">
        <button
          v-for="b in item.branches" :key="b.name"
          class="text-xs px-2 py-1 rounded border transition-colors"
          :class="selections[item.name] === b.name
            ? (b.create_todo ? 'bg-red-500 text-white border-red-500' : 'bg-green-500 text-white border-green-500')
            : 'border-gray-200 dark:border-gray-600 text-gray-500 dark:text-gray-400 hover:border-primary'"
          @click="selectBranch(item.name, b.name)"
        >{{ b.name }}</button>
      </div>
    </div>
  </div>
</template>
